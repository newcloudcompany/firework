#[macro_use]
extern crate log;

use std::{env, fs};

use std::fs::File;
use std::io::{self, Read, Write};
use std::os::unix::io::{AsRawFd, FromRawFd};

use std::process::Command;
use std::sync::mpsc;

use anyhow::Error;
// use futures::{StreamExt, TryFutureExt, TryStreamExt};
use ptyca::{openpty, PtyCommandExt};

use nix::errno::Errno;

use nix::mount::{mount as nix_mount, MsFlags};

use nix::sys::{
    stat::Mode,
    wait::{waitpid, WaitPidFlag, WaitStatus},
};
use nix::unistd::{mkdir as nix_mkdir, symlinkat};
use nix::NixPath;

use vsock::{VsockListener, VsockStream};

pub fn log_init() {
    // default to "info" level, just for this bin
    let level = env::var("LOG_FILTER").unwrap_or_else(|_| "init=info".into());

    env_logger::builder()
        .parse_filters(&level)
        .write_style(env_logger::WriteStyle::Never)
        .format_module_path(false)
        .init();
}

fn main() -> Result<(), InitError> {
    log_init();

    // can't put these as const unfortunately...
    let chmod_0755: Mode =
        Mode::S_IRWXU | Mode::S_IRGRP | Mode::S_IXGRP | Mode::S_IROTH | Mode::S_IXOTH;
    let chmod_0555: Mode = Mode::S_IRUSR
        | Mode::S_IXUSR
        | Mode::S_IRGRP
        | Mode::S_IXGRP
        | Mode::S_IROTH
        | Mode::S_IXOTH;
    let chmod_1777: Mode = Mode::S_IRWXU | Mode::S_IRWXG | Mode::S_IRWXO | Mode::S_ISVTX;
    // let chmod_0777 = Mode::S_IRWXU | Mode::S_IRWXG | Mode::S_IRWXO;
    let common_mnt_flags: MsFlags = MsFlags::MS_NODEV | MsFlags::MS_NOEXEC | MsFlags::MS_NOSUID;

    info!("Starting init...");

    info!("Mounting /dev/pts");
    mkdir("/dev/pts", chmod_0755).ok();
    mount(
        Some("devpts"),
        "/dev/pts",
        Some("devpts"),
        MsFlags::MS_NOEXEC | MsFlags::MS_NOSUID | MsFlags::MS_NOATIME,
        Some("mode=0620,gid=5,ptmxmode=666"),
    )?;

    info!("Mounting /dev/mqueue");
    mkdir("/dev/mqueue", chmod_0755).ok();
    mount::<_, _, _, [u8]>(
        Some("mqueue"),
        "/dev/mqueue",
        Some("mqueue"),
        common_mnt_flags,
        None,
    )?;

    info!("Mounting /dev/shm");
    mkdir("/dev/shm", chmod_1777).ok();
    mount::<_, _, _, [u8]>(
        Some("shm"),
        "/dev/shm",
        Some("tmpfs"),
        MsFlags::MS_NOSUID | MsFlags::MS_NODEV,
        None,
    )?;

    info!("Mounting /dev/hugepages");
    mkdir("/dev/hugepages", chmod_0755).ok();
    mount(
        Some("hugetlbfs"),
        "/dev/hugepages",
        Some("hugetlbfs"),
        MsFlags::MS_RELATIME,
        Some("pagesize=2M"),
    )?;

    info!("Mounting /proc");
    mkdir("/proc", chmod_0555).ok();
    mount::<_, _, _, [u8]>(Some("proc"), "/proc", Some("proc"), common_mnt_flags, None)?;
    mount::<_, _, _, [u8]>(
        Some("binfmt_misc"),
        "/proc/sys/fs/binfmt_misc",
        Some("binfmt_misc"),
        common_mnt_flags | MsFlags::MS_RELATIME,
        None,
    )?;

    info!("Mounting /sys");
    mkdir("/sys", chmod_0555).ok();
    mount::<_, _, _, [u8]>(Some("sys"), "/sys", Some("sysfs"), common_mnt_flags, None)?;

    info!("Mounting /run");
    mkdir("/run", chmod_0755).ok();
    mount(
        Some("run"),
        "/run",
        Some("tmpfs"),
        MsFlags::MS_NOSUID | MsFlags::MS_NODEV,
        Some("mode=0755"),
    )?;
    mkdir("/run/lock", Mode::all()).ok();

    symlinkat("/proc/self/fd", None, "/dev/fd").ok();
    symlinkat("/proc/self/fd/0", None, "/dev/stdin").ok();
    symlinkat("/proc/self/fd/1", None, "/dev/stdout").ok();
    symlinkat("/proc/self/fd/2", None, "/dev/stderr").ok();

    mkdir("/root", Mode::S_IRWXU).ok();

    let common_cgroup_mnt_flags =
        MsFlags::MS_NODEV | MsFlags::MS_NOEXEC | MsFlags::MS_NOSUID | MsFlags::MS_RELATIME;

    info!("Mounting cgroup");
    mount(
        Some("tmpfs"),
        "/sys/fs/cgroup",
        Some("tmpfs"),
        MsFlags::MS_NOSUID | MsFlags::MS_NOEXEC | MsFlags::MS_NODEV, // | MsFlags::MS_RDONLY,
        Some("mode=755"),
    )?;

    info!("Mounting cgroup2");
    mkdir("/sys/fs/cgroup/unified", chmod_0555)?;
    mount(
        Some("cgroup2"),
        "/sys/fs/cgroup/unified",
        Some("cgroup2"),
        common_mnt_flags | MsFlags::MS_RELATIME,
        Some("nsdelegate"),
    )?;

    info!("Mounting /sys/fs/cgroup/net_cls,net_prio");
    mkdir("/sys/fs/cgroup/net_cls,net_prio", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/net_cls,net_prio",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("net_cls,net_prio"),
    )?;

    info!("Mounting /sys/fs/cgroup/hugetlb");
    mkdir("/sys/fs/cgroup/hugetlb", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/hugetlb",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("hugetlb"),
    )?;

    info!("Mounting /sys/fs/cgroup/pids");
    mkdir("/sys/fs/cgroup/pids", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/pids",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("pids"),
    )?;

    info!("Mounting /sys/fs/cgroup/freezer");
    mkdir("/sys/fs/cgroup/freezer", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/freezer",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("freezer"),
    )?;

    info!("Mounting /sys/fs/cgroup/cpu,cpuacct");
    mkdir("/sys/fs/cgroup/cpu,cpuacct", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/cpu,cpuacct",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("cpu,cpuacct"),
    )?;

    info!("Mounting /sys/fs/cgroup/devices");
    mkdir("/sys/fs/cgroup/devices", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/devices",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("devices"),
    )?;

    info!("Mounting /sys/fs/cgroup/blkio");
    mkdir("/sys/fs/cgroup/blkio", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/blkio",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("blkio"),
    )?;

    info!("Mounting cgroup/memory");
    mkdir("/sys/fs/cgroup/memory", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/memory",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("memory"),
    )?;

    info!("Mounting /sys/fs/cgroup/perf_event");
    mkdir("/sys/fs/cgroup/perf_event", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/perf_event",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("perf_event"),
    )?;

    info!("Mounting /sys/fs/cgroup/cpuset");
    mkdir("/sys/fs/cgroup/cpuset", chmod_0555)?;
    mount(
        Some("cgroup"),
        "/sys/fs/cgroup/cpuset",
        Some("cgroup"),
        common_cgroup_mnt_flags,
        Some("cpuset"),
    )?;

    // rlimit::setrlimit(rlimit::Resource::NOFILE, 10240, 10240).ok();

    mkdir("/etc", chmod_0755).ok();

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

    let cid = client
        .get(&format!("http://{}/cid", addr))
        .header("X-metadata-token", token)
        .send()
        .expect("failed to send")
        .text()
        .expect("failed to text");

    info!("cid: {}", cid);

    // Make cid u32.
    let cid = cid.parse::<u32>().expect("failed to convert");

    // Enable packet forwarding.
    fs::write("/proc/sys/net/ipv4/conf/all/forwarding", "1")?;

    // Set standard PATH env variable.
    env::set_var(
        "PATH",
        "/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin",
    );

    let listener = VsockListener::bind_with_cid_port(cid, 10000).expect("failed to bind vsock");
    for stream in listener.incoming() {
        std::thread::spawn(|| {
            handle_conn(stream.expect("bad connection")).expect("failed handling connection")
        });
    }

    Ok(())
}

fn reap_zombies(pid: i32, exit_status: &mut i32) -> bool {
    let mut child_exited = false;
    loop {
        match waitpid(None, Some(WaitPidFlag::WNOHANG)) {
            Ok(status) => {
                if Some(pid) == status.pid().map(nix::unistd::Pid::as_raw) {
                    // main process pid exited
                    child_exited = true;
                }
                match status {
                    WaitStatus::Exited(child_pid, exit_code) => {
                        if child_pid.as_raw() == pid {
                            info!("Main child exited normally with code: {}", exit_code);
                            *exit_status = exit_code;
                        } else {
                            warn!(
                                "Reaped child process with pid: {}, exit code: {}",
                                child_pid, exit_code
                            )
                        }
                    }
                    WaitStatus::Signaled(child_pid, signal, core_dumped) => {
                        if child_pid.as_raw() == pid {
                            info!(
                                "Main child exited with signal (with signal '{}', core dumped? {})",
                                signal, core_dumped
                            );
                            *exit_status = 128 + (signal as i32);
                        } else {
                            warn!(
                                "Reaped child process with pid: {} and signal: {}, core dumped? {}",
                                child_pid, signal, core_dumped
                            )
                        }
                    }
                    WaitStatus::Stopped(child_pid, signal) => {
                        info!(
                            "waitpid Stopped: surprising (pid: {}, signal: {})",
                            child_pid, signal
                        );
                    }
                    WaitStatus::PtraceEvent(child_pid, signal, event) => {
                        info!(
                            "waitpid PtraceEvent: interesting (pid: {}, signal: {}, event: {})",
                            child_pid, signal, event
                        );
                    }
                    WaitStatus::PtraceSyscall(child_pid) => {
                        info!("waitpid PtraceSyscall: unfathomable (pid: {})", child_pid);
                    }
                    WaitStatus::Continued(child_pid) => {
                        info!("waitpid Continue: not supposed to! (pid: {})", child_pid);
                    }
                    WaitStatus::StillAlive => {
                        trace!("no more children to reap");
                        break;
                    }
                }
            }
            Err(e) => match e {
                Errno::ECHILD => {
                    info!("no child to wait");
                    break;
                }
                Errno::EINTR => {
                    info!("got EINTR waiting for pids, continuing...");
                    continue;
                }
                _ => {
                    info!("error calling waitpid: {}", e);
                    // TODO: return an error? handle it?
                    return false;
                }
            },
        }
    }
    child_exited
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
    let mut reader = writer.try_clone()?;
    println!("Incoming connection from {}...", writer.peer_addr()?);

    let (primary, secondary) = openpty()?;
    let primary_raw_fd = primary.as_raw_fd();

    let mut cmd = Command::new("bash");
    cmd.arg("-i");
    cmd.envs(env::vars());

    let mut child = cmd.spawn_pty(secondary)?;

    let (tx, rx) = mpsc::channel();
    let primary_read_tx = tx.clone();

    let mut primary = unsafe { File::from_raw_fd(primary_raw_fd) };
    let mut primary_clone = primary.try_clone()?;

    std::thread::spawn(move || {
        let mut buffer = vec![0u8; 1024];
        loop {
            match primary_clone.read(&mut buffer) {
                Ok(n) if n > 0 => {
                    let slice = &buffer[..n];
                    let _ = primary_read_tx.send(Msg::PrimaryRead(slice.to_vec()));
                }
                _ => break,
            }
        }
    });

    std::thread::spawn(move || {
        let _ = child.wait();
        let _ = tx.send(Msg::Exit);
    });

    std::thread::spawn(move || {
        let _ = io::copy(&mut reader, &mut primary);
    });

    loop {
        let msg = rx.recv()?;
        match msg {
            Msg::PrimaryRead(buffer) => {
                let _ = writer.write_all(&buffer);
                let _ = writer.flush();
            }
            Msg::Exit => break,
        }
    }

    println!("Closing connection from {}...", writer.peer_addr()?);
    writer.shutdown(std::net::Shutdown::Both)?; // TODO: Maybe properly join the writer to primary instead.
    Ok(())
}
