# i3-go-quickterm

A small drop-down terminal for [i3wm](https://i3wm.org/) and [sway](https://swaywm.org/).

This is a Golang rewrite version of [i3-quickterm](https://github.com/lbonn/i3-quickterm) for self easy install and use.

## Prerequisites

[termite](https://wiki.archlinux.org/index.php/Termite) used as default terminal emulator, so install it firstly.

## Install

```
# install form source code
$ go get -u github.com/zexi/i3-go-quickterm

# run quickterm inside sway or i3 WM
$ $GOPATH/bin/i3-go-quickterm
```

## Configuration

The configuration is read from ~/.config/i3-go-quickterm/config.json .

```bash
$ cat $HOME/.config/i3-go-quickterm/config.json
{
  "terminal": {
    "command": "termite",
    "extraArgs": [
      "-c",
      "/home/lzx/.config/termite/config.trans"
    ]
  },
  "pos": "top",
  "ratio": 0.5
}
```

sway or i3 config

```bash
# always pop drop down shell
bindsym $mod+g exec "$GOPATH/bin/i3-go-quickterm"
```
