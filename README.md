![screenshot of mpq](./screenshot.png)

With mpq you can view and manipulate songs in the mpd queue. These are
the default key bindings; customize by modifying `keys.go`:

```console
$ mpq -h
Key bindings:
q       : quit
enter   : play highlighted song
space   : toggle play/pause
up      : highlight previous song
down    : highlight next song
alt-up  : move highlighted song up
alt-down: move highlighted song down
left    : seek backwards 5s
right   : seek forwards 5s
d       : remove song from queue
```

# Installation
```shell
git clone git@github.com:codesoap/mpq.git
cd mpq
go install

# Now you can run mpq:
mpq
```

# Adding songs to the queue
I use a little script called `mqa` (music queue add). it requires `mpc`
and `fzf`; use `tab` to select songs and `enter` to add them to the
queue:

```shell
#!/usr/bin/env sh

alias mpc='mpc -f "%file%\t[%artist% - ][%album% [#[##%track%#] ]- ][%title%|%file%]"'
songs=$(mpc listall | sort -V)
selections=$(printf '%s\n' "$songs" | awk -F'\t' '{print $2}' | fzf --no-sort --reverse -m)
selected_uris=$(
	printf '%s\n' "$selections" | while read selection
	do
		printf '%s\n' "$songs" | grep -F "$selection" | awk -F'\t' '{print $1; exit}'
	done
)
printf '%s\n' "$selected_uris" | mpc add
```

# Configuring mpd
Other tasks, like disabling the repeat mode, I just do with `mpc`. My
config is usually this:

```shell
mpc consume on
mpc crossfade 0
mpc random off
mpc repeat off
mpc replaygain album
mpc single off
mpc volume 100
```
