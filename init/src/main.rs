use std::{
    env, io,
    process::{Command, Stdio},
    sync::mpsc::{self},
    thread::{self},
};

use nix::{
    libc::{c_int, pid_t},
    sys::{
        signal::Signal,
        wait::{waitpid, WaitStatus},
    },
    unistd::Pid,
};

use log::{debug, error};
use nix::sys::signal::kill;
use signal_hook::consts::signal::SIGINT;

pub fn log_init() {
    // default to "debug" level, just for this bin
    let level = env::var("LOG_FILTER").unwrap_or_else(|_| "init=debug".into());

    env_logger::builder()
        .parse_filters(&level)
        .write_style(env_logger::WriteStyle::Never)
        .format_module_path(false)
        .init();
}

fn notify(signals: &[c_int]) -> anyhow::Result<mpsc::Receiver<c_int>> {
    let (tx, rx) = mpsc::channel();
    let mut signals = signal_hook::iterator::Signals::new(signals)?;

    thread::spawn(move || {
        for signal in signals.forever() {
            if tx.send(signal).is_err() {
                break;
            }
        }
    });

    Ok(rx)
}

fn reap_zombies(child_pid: pid_t) -> i32 {
    debug!("Starting loop that reaps zombies");

    loop {
        match waitpid(Pid::from_raw(-1), None) {
            Ok(status) => match status {
                WaitStatus::Exited(pid, code) => {
                    if pid.as_raw() == child_pid {
                        debug!("Main child (pid: {pid}) exited with code {code}");
                        return code;
                    } else {
                        debug!("Reaped child (pid: {pid}). Exit code: {code}");
                    }
                }
                WaitStatus::Signaled(pid, signal, _) => {
                    if pid.as_raw() == child_pid {
                        debug!("Main child exited with signal (with signal: {signal:?})");
                        return 128 + (signal as i32);
                    } else {
                        debug!("Reaped child process (pid: {pid}, signal: {signal:?})")
                    }
                }
                status => {
                    debug!("waitpid returned: {status:?}");
                }
            },
            Err(e) => {
                error!("Error: {}", e);
                return 1;
            }
        }
    }
}

fn main() -> Result<(), anyhow::Error> {
    log_init();

    let exe = env::args().nth(1).unwrap_or_else(|| {
        error!("Provide a path to an executable as an argument");
        log::logger().flush();
        std::process::exit(1);
    });

    let mut cmd = Command::new(exe);
    cmd.stdout(io::stdout())
        .stderr(io::stderr())
        .stdin(Stdio::inherit());

    let child = cmd.spawn()?;
    let child_pid = child.id() as pid_t;
    debug!("Spawned child, pid: {}", child.id());

    let signals = notify(&[SIGINT])?;
    debug!("Start loop that intercepts signals");

    thread::spawn(move || {
        for sig in signals {
            let signal = Signal::try_from(sig).expect("Signal is not valid");
            debug!("Init received signal: {signal}");

            match kill(Pid::from_raw(child_pid), signal) {
                Ok(_) => {
                    debug!("Forwarded {signal} to the child");
                    break;
                }
                Err(e) => {
                    error!("Error: {}", e);
                }
            }
        }
    });

    let exit_code = reap_zombies(child_pid);
    std::process::exit(exit_code);
}
