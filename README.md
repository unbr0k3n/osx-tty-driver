# OSX TTY Driver installer

OSX ships with a terminal driver that supports the framebuffer extension, but it needs to be signed in order to activate it.

This is a small Golang wrapper providing the proper AES signature to enable the driver. Especially useful if we wrap [ncurses](https://www.gnu.org/software/ncurses/) or any other TUI library in Go that requires fancy graphics in the browser

## Installation

```sh
$ go get "github.com/unbr0k3n/osx-tty-driver"
```

## Usage

```go
import (
    ...
    tty "github.com/unbr0k3n/osx-tty-driver"
)

func main() {
    // ensure driver is signed, no-op on non-osx platforms or if driver is already signed.
    tty.Ensure()
}

```