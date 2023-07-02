use rustix::fd::OwnedFd;
use rustix::pty::{grantpt, openpt, ptsname, unlockpt, OpenptFlags};
use rustix::termios::{tcgetattr, tcsetattr};
use rustix::{
    fs::{cwd, openat, Mode},
    termios::{cfmakeraw, OptionalActions, Termios},
};
use std::io;
use std::{os::unix::process::CommandExt, process::Command};

use rustix::process::setsid;

use rustix::{fd::AsRawFd, io::close, process::ioctl_tiocsctty};

pub fn openpty() -> Result<(OwnedFd, OwnedFd), Box<dyn std::error::Error>> {
    let flags = OpenptFlags::RDWR | OpenptFlags::NOCTTY | OpenptFlags::CLOEXEC;
    let primary = openpt(flags)?;

    grantpt(&primary)?;
    unlockpt(&primary)?;

    let secondary_name = ptsname(&primary, Vec::new())?;
    let secondary = openat(cwd(), secondary_name, flags.into(), Mode::empty())?;

    Ok((primary, secondary))
}

pub fn make_terminal_raw() -> Result<Termios, Box<dyn std::error::Error>> {
    let attrs = tcgetattr(io::stdin())?;
    let mut attrs_copy = attrs;
    cfmakeraw(&mut attrs_copy);
    tcsetattr(io::stdin(), OptionalActions::Now, &attrs_copy)?;
    Ok(attrs)
}

pub fn restore_terminal(attrs: &Termios) -> Result<(), Box<dyn std::error::Error>> {
    tcsetattr(io::stdin(), OptionalActions::Now, attrs)?;
    Ok(())
}

pub trait PtyCommandExt {
    fn spawn_pty(
        &mut self,
        secondary: OwnedFd,
    ) -> Result<std::process::Child, Box<dyn std::error::Error>>;
}

impl PtyCommandExt for Command {
    fn spawn_pty(
        &mut self,
        secondary: OwnedFd,
    ) -> Result<std::process::Child, Box<dyn std::error::Error>> {
        let secondary_raw_fd = secondary.as_raw_fd();

        self.stdin(secondary.try_clone()?);
        self.stdout(secondary.try_clone()?);
        self.stderr(secondary.try_clone()?);

        unsafe {
            self.pre_exec(move || {
                setsid()?;
                ioctl_tiocsctty(&secondary)?;
                Ok(())
            });
        }

        let child = self.spawn()?;
        unsafe {
            close(secondary_raw_fd);
        }

        Ok(child)
    }
}
