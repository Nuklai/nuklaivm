# Nuklai Wallet

## Live Development

To run in live development mode, run `./scripts/dev.sh` in the project directory. This will do the following:

- Build `nuklai-faucet`, `nuklai-feed` and `nuklai-wallet`
- Run Nuklai Faucet with the default settings by connecting to the local Nuklai network
- Run Nuklai Feed with the default settings by connecting to the local Nuklai network
- Run Nuklai Wallet with the default settings by connecting to Nuklai Faucet, Nuklai Feed and the local Nuklai network.

The Nuklai Wallet is a Vite development server that provides very fast hot reload of your frontend changes.

By default, it'll open a new desktop app for the wallet however, you can also go to [http://localhost:34115](http://localhost:34115) on your browser to interact with the wallet.

## Release

To build a distributable package for MacOS, run `./scripts/build.sh`.

## Debug

When running `./scripts/dev.sh`, if you see an error like:

```
# pkg-config --cflags  -- gtk+-3.0 webkit2gtk-4.0 gio-unix-2.0 gtk+-3.0 webkit2gtk-4.0 gio-unix-2.0 gtk+-3.0 webkit2gtk-4.0
Package gtk+-3.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gtk+-3.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'gtk+-3.0' found
Package webkit2gtk-4.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `webkit2gtk-4.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'webkit2gtk-4.0' found
Package gio-unix-2.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gio-unix-2.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'gio-unix-2.0' found
Package gtk+-3.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gtk+-3.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'gtk+-3.0' found
Package webkit2gtk-4.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `webkit2gtk-4.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'webkit2gtk-4.0' found
Package gio-unix-2.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gio-unix-2.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'gio-unix-2.0' found
Package gtk+-3.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `gtk+-3.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'gtk+-3.0' found
Package webkit2gtk-4.0 was not found in the pkg-config search path.
Perhaps you should add the directory containing `webkit2gtk-4.0.pc'
to the PKG_CONFIG_PATH environment variable
No package 'webkit2gtk-4.0' found
pkg-config: exit status 1

Build error - exit status 1
```

You should probably install the required packages in your system.

On Ubuntu/Debian-based systems:

```bash
sudo apt-get update
sudo apt-get install -y libgtk-3-dev libwebkit2gtk-4.0-dev libgirepository1.0-dev
```

On Fedora:

```bash
sudo dnf install -y gtk3-devel webkit2gtk3-devel gobject-introspection-devel
```

On Arch Linux:

```bash
sudo pacman -Syu gtk3 webkit2gtk gobject-introspection
```

On Mac:

```
sudo port install webkit2-gtk
```

Then, try running `./scripts/dev.sh` again

If you run the command `./scripts/dev.sh` in a vscode terminal and encounter the error `symbol lookup error: /snap/core20/current/lib/x86_64-linux-gnu/libpthread.so.0: undefined symbol: __libc_pthread_init, version GLIBC_PRIVATE`, you should run the command in a console.
