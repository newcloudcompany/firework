#[macro_use]
extern crate log;

use std::collections::HashMap;
use std::{env, fs};

use std::fs::File;
use std::io::{self, Read, Write};
use std::os::unix::io::{AsRawFd, FromRawFd};

use std::process::Command;
use std::sync::mpsc;

use anyhow::Error;
use nom::bytes::complete::tag;
use nom::sequence::{tuple, terminated, preceded};
use nom::number::Endianness;
use ptyca::{openpty, PtyCommandExt};

use nix::errno::Errno;

use nix::mount::{mount as nix_mount, MsFlags};

use nix::sys::{
    stat::Mode,
    wait::{waitpid, WaitPidFlag, WaitStatus},
};
use nix::unistd::{mkdir as nix_mkdir, symlinkat};
use nix::NixPath;

use rustix::fs::MountFlags;
use rustix::system::sethostname;
use rustix::termios::{tcsetwinsize, Winsize};
use serde::Deserialize;
use vsock::{VsockListener, VsockStream};

#[derive(Deserialize)]
    struct Metadata {
        cid: u32,
        ipv4: String,
        hostname: String,
        hosts: HashMap<String, String>,
    }

pub fn log_init() {
    // default to "info" level, just for this bin
    let level = env::var("LOG_FILTER").unwrap_or_else(|_| "init=info".into());

    env_logger::builder()
        .parse_filters(&level)
        .write_style(env_logger::WriteStyle::Never)
        .format_module_path(false)
        .init();
}

fn main() -> Result<(), anyhow::Error> {
    log_init();

    // Use HTTP client (reqwest) to call Firecracker MMDS endpoint to retrieve metadata that contains the CID for the vsock listener.
    // First it must call the token endpoint (/latest/api/token) with PUT method and X-metadata-token-ttl-seconds header to issue a session token.
    // Then the token is used in the X-metadata-token header to make a call to latest/meta-data endpoint.
    // MMDS IPv4 address: 169.254.169.254

    let addr = "169.254.169.254";

    // Add route to MMDS.
    let mut cmd = Command::new("ip");
    let output = cmd.args(["route", "add", addr, "dev", "eth0"]).output()?;
    println!("{}", String::from_utf8_lossy(&output.stderr));
    println!("{}", String::from_utf8_lossy(&output.stdout));

    let client = reqwest::blocking::Client::new();
    let token = client
        .put(&format!("http://{}/latest/api/token", addr))
        .header("X-metadata-token-ttl-seconds", "21600")
        .send()
        .expect("failed to send")
        .text()
        .expect("failed to text");

    info!("token: {}", token);

    let metadata = client
        .get(&format!("http://{}", addr))
        .header("X-metadata-token", token)
        .header("Accept", "application/json")
        .send()
        .expect("failed to send")
        .json::<Metadata>().expect("failed to json");

    // Enable packet forwarding.
    fs::write("/proc/sys/net/ipv4/conf/all/forwarding", "1")?;

    sethostname(metadata.hostname.as_bytes()).expect("failed to set hostname");

    let hosts_string = metadata.hosts.iter()
        .map(|(k, v)| format!("{} {}", v, k))
        .collect::<Vec<String>>().join("\n");

    fs::write("/etc/hosts", hosts_string)?;
    
    // Set standard PATH env variable.
    env::set_var(
        "PATH",
        "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    );

    let listener = VsockListener::bind_with_cid_port(metadata.cid, 10000).expect("failed to bind vsock");
    for stream in listener.incoming() {
        std::thread::spawn(|| {
            handle_conn(stream.expect("bad connection")).expect("failed handling connection")
        });
    }

    Ok(())
}

#[derive(Debug, thiserror::Error)]
enum InitError {
    #[error("couldn't mount {} onto {}, because: {}", source, target, error)]
    Mount {
        source: String,
        target: String,
        #[source]
        error: nix::Error,
    },

    #[error("couldn't mkdir {}, because: {}", path, error)]
    Mkdir {
        path: String,
        #[source]
        error: nix::Error,
    },

    #[error("an unhandled error occurred: {}", 0)]
    UnhandledNixError(#[from] nix::Error),

    #[error("an unhandled IO error occurred: {}", 0)]
    UnhandledIoError(#[from] io::Error),

    #[error("an unhandled error occurred: {}", 0)]
    UnhandledError(#[from] Error),
}

use nom::character::complete::u16;

fn try_parse_resize_msg(input: &[u8]) -> nom::IResult<&[u8], (u16, u16)> {
    preceded(tag("RESIZE,"), tuple((
        terminated(u16, tag(",")),
        terminated(u16, tag(","))
    )))(input)
}

#[test]
fn test_parse_resize_msg() {
    assert_eq!(try_parse_resize_msg(b"RESIZE,80,24,"), Ok((&[][..], (80, 24))));
}

fn mount<P1: ?Sized + NixPath, P2: ?Sized + NixPath, P3: ?Sized + NixPath, P4: ?Sized + NixPath>(
    source: Option<&P1>,
    target: &P2,
    fstype: Option<&P3>,
    flags: MsFlags,
    data: Option<&P4>,
) -> Result<(), InitError> {
    nix_mount(source, target, fstype, flags, data).map_err(|error| InitError::Mount {
        source: source
            .map(|p| {
                p.with_nix_path(|cs| cs.to_owned().into_string().ok().unwrap_or_default())
                    .unwrap_or_else(|_| String::new())
            })
            .unwrap_or_else(String::new),
        target: target
            .with_nix_path(|cs| cs.to_owned().into_string().ok().unwrap_or_default())
            .unwrap_or_else(|_| String::new()),
        error,
    })
}

fn mkdir<P: ?Sized + NixPath>(path: &P, mode: Mode) -> Result<(), InitError> {
    nix_mkdir(path, mode).map_err(|error| InitError::Mkdir {
        path: path
            .with_nix_path(|cs| cs.to_owned().into_string().ok().unwrap_or_default())
            .unwrap_or_else(|_| String::new()),
        error,
    })
}

enum Msg {
    PrimaryRead(Vec<u8>),
    Exit,
}

fn handle_conn(mut writer: VsockStream) -> Result<(), Box<dyn std::error::Error>> {
    info!("Incoming connection from {}", writer.peer_addr()?);

    let reader = writer.try_clone()?;

    let (primary, secondary) = openpty()?;
    let primary_raw_fd = primary.as_raw_fd();

    let mut cmd = Command::new("bash");
    cmd.arg("-i");
    cmd.envs(env::vars());

    let mut child = cmd.spawn_pty(secondary)?;

    let (tx, rx) = mpsc::channel();
    let primary_read_tx = tx.clone();

    let primary = unsafe { File::from_raw_fd(primary_raw_fd) };
    let primary_clone = primary.try_clone()?;

    let conn_reader = std::thread::spawn(move || copy_primary_to_conn(primary_clone, primary_read_tx));
    let primary_reader = std::thread::spawn(move || copy_conn_to_primary(reader, primary));
    let child_waiter = std::thread::spawn(move || {
        let _ = child.wait();
        let _ = tx.send(Msg::Exit);
    });

    loop {
        let msg = rx.recv()?;
        match msg {
            Msg::PrimaryRead(buffer) => {
                writer.write_all(&buffer)?;
                writer.flush()?;
            }
            Msg::Exit => break,
        }
    }

    writer.shutdown(std::net::Shutdown::Both)?;

    let _ = child_waiter.join().or(Err("Failed to join child_waiter thread"))?;
    let _ = conn_reader.join().or(Err("Failed to join conn_reader thread"))?;
    let _ = primary_reader.join().or(Err("Failed to join primary_reader thread"))?;

    info!("Closed connection from {}", writer.peer_addr()?);
    Ok(())
}

fn copy_conn_to_primary(mut conn: VsockStream, mut primary: File) -> Result<(), anyhow::Error> {
    let mut buffer = vec![0u8; 1024];
    while let Ok(n) =  conn.read(&mut buffer) {
        if n == 0 { break }

        let slice = &buffer[..n];
        match try_parse_resize_msg(slice) {
            Ok((_, (width, height))) => {
                let ws = Winsize {
                    ws_row: height,
                    ws_col: width,
                    ws_xpixel: 0,
                    ws_ypixel: 0,
                };

                tcsetwinsize(&primary, ws)?
            },
            _ =>  primary.write_all(slice)?
        }
    }
    Ok(())
}

fn copy_primary_to_conn(mut primary: File, tx: mpsc::Sender<Msg>) -> Result<(), anyhow::Error> {
    let mut buffer = vec![0u8; 1024];
    while let Ok(n) =  primary.read(&mut buffer) {
        if n == 0 { break }
        let slice = &buffer[..n];
        let _ = tx.send(Msg::PrimaryRead(slice.to_vec()));
    }
    Ok(())
}
