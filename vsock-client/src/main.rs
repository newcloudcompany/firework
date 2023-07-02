use ptyca::{make_terminal_raw, restore_terminal};
use std::{
    io::{self, Read, Write},
    os::unix::net::UnixStream,
};

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let path = std::env::args()
        .nth(1)
        .unwrap_or_else(|| panic!("no path given"));
    let old_state = make_terminal_raw()?;

    let mut writer = UnixStream::connect(path)?;

    // writer.set_nodelay(true)?;
    let mut reader = writer.try_clone()?;
    writer.write_all("CONNECT 10000\n".as_bytes())?;

    std::thread::spawn(move || {
        let _ = io::copy(&mut io::stdin(), &mut writer);
    });

    let mut buffer = vec![0u8; 1024];
    loop {
        match reader.read(&mut buffer) {
            Ok(n) if n > 0 => {
                let slice = &buffer[..n];
                let _ = io::stdout().write_all(slice);
                let _ = io::stdout().flush();
            }
            _ => break,
        }
    }

    println!("Reader exited.");
    // let _ = stdout().write_all(&buffer[..]);
    let _ = io::stdout().flush();

    restore_terminal(&old_state)?;

    Ok(())
}
