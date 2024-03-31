![screenshot of mpq](./screenshot.png)

With mpq you can view and manipulate songs in the mpd queue. These are
the default key bindings; customize by modifying `keys.go`:

```console
$ mpq -h
Key bindings:
q          : quit
enter      : play highlighted song
space      : toggle play/pause
up,k       : highlight previous song
down,j     : highlight next song
alt+up/k   : move highlighted song up
alt+down/j : move highlighted song down
left,h     : seek backwards 5s
right,l    : seek forwards 5s
d          : remove song from queue
c          : clear queue
```

# Installation
```shell
git clone git@github.com:codesoap/mpq.git
cd mpq
go install

# The binary is now at ~/go/bin/mpq. Add ~/go/bin to your $PATH to run
# go programs easily:
mpq
```

# Adding songs to the queue
I use a little script called `mqa` (music queue add). It requires `mpc`
and `fzf`; use `tab` to select songs and `enter` to add them to the
queue:

```shell
#!/usr/bin/env sh

alias mpc='mpc -f "%file%\t[%artist% - ][%album% [#[##%track%#] ]- ][%title%|%file%]"'
songs=$(mpc listall | sort -V)
printf '%s\n' "$songs" \
| awk -F'\t' '{print $2}' \
| fzf --no-sort --reverse -m \
| while read selection
do
	printf '%s\n' "$songs" | awk -F'\t' "\$2==\"$selection\" {print \$1; exit}"
done | mpc add
```

## Utilizing songmem
[songmem](https://github.com/codesoap/songmem/) is another tool I wrote
to store and analyze the songs I listen to. It can be used, for example,
to find recommendations for the last heard song and add them to the
queue with a script like this:

```shell
#!/usr/bin/env sh

selection=$(songmem --suggestions "$(songmem | head -n1)" | fzf)
artist="$(printf '%s' "$selection" | awk -F ' - ' '{print $1}')"
title="$(printf '%s' "$selection" | awk '{i=index($0, " - "); print substr($0, i+3)}')"
mpc findadd artist "$artist" title "$title"
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
