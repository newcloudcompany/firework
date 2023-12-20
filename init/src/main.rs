use std::{
    env, io,
    process::{Command, Stdio},
    sync::mpsc::{self, channel},
    thread::{self, sleep},
    time::Duration,
};

use nix::{
    libc::{c_int, pid_t},
    sys::{
        signal::Signal,
        wait::{waitpid, WaitPidFlag, WaitStatus},
    },
    unistd::Pid,
};

use log::{debug, error};

pub fn log_init() {
    // default to "debug" level, just for this bin
    let level = env::var("LOG_FILTER").unwrap_or_else(|_| "init=debug".into());

    env_logger::builder()
        .parse_filters(&level)
        .write_style(env_logger::WriteStyle::Never)
        .format_module_path(false)
        .init();
}

fn main() -> Result<(), anyhow::Error> {
    // let exe = env::args().nth(1).unwrap_or_else(|| {
    //     eprintln!("Please provide a path to an executable");
    //     std::process::exit(1);
    // });
    log_init();

    let mut cmd = Command::new("sleep");
    cmd.arg("10");
    cmd.stdout(io::stdout())
        .stderr(io::stderr())
        .stdin(Stdio::inherit());

    let child = cmd.spawn()?;
    debug!("Spawned child (pid: {})", child.id());
    use signal_hook::consts::signal::SIGINT;

    let signals = notify(&[SIGINT])?;

    use nix::sys::signal::kill;
    let child_pid = child.id() as pid_t;

    thread::spawn(move || {
        for sig in signals {
            let signal = Signal::try_from(sig).expect("bad signal");
            debug!("Received signal: {signal}");
            match kill(Pid::from_raw(child_pid), signal) {
                Ok(_) => {
                    debug!("Sent signal {signal} to the child");
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
    loop {
        match waitpid(Pid::from_raw(-1), None) {
            Ok(status) => match status {
                WaitStatus::Exited(pid, code) => {
                    if pid.as_raw() == child_pid {
                        debug!("Main child (pid {pid}) exited with code {code}");
                        return code;
                    } else {
                        debug!("Reaped child (pid {pid}). Exit code: {code}");
                    }
                }
                WaitStatus::Signaled(pid, signal, _) => {
                    if pid.as_raw() == child_pid {
                        debug!("Main child exited with signal (with signal '{signal:?}')");
                        return 128 + (signal as i32);
                    } else {
                        debug!("Reaped child process (pid: {pid}, signal: {signal:?})")
                    }
                }
                _ => {
                    debug!("Child still running");
                }
            },
            Err(e) => {
                error!("Error: {}", e);
                return 1;
            }
        }
    }
}
