<p align="center">
  <img src="https://github.com/macaroni-os/macaroni-site/blob/master/site/static/images/logo.png">
</p>

# Macaroni OS System Management Tool

[![Build on push](https://github.com/macaroni-os/macaronictl/actions/workflows/push.yml/badge.svg)](https://github.com/macaroni-os/macaronictl/actions/workflows/push.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/macaroni-os/macaronictl)](https://goreportcard.com/report/github.com/macaroni-os/macaronictl)
[![CodeQL](https://github.com/macaroni-os/macaronictl/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/macaroni-os/macaronictl/actions/workflows/codeql-analysis.yml)

The Macaroni OS knife tool to control your system.

At the moment, it contains only the commands to control
the kernels and generate initrd images.


```
$ macaronictl --help
Copyright (c) 2020-2023 Macaroni OS - Daniele Rondina

Macaroni Linux System Management Tool

Usage:
   [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  env-update  Updates environment settings automatically.
  etc-update  Handle configuration file updates.
  help        Help about any command
  kernel      Manage system kernels and initrd.

Flags:
  -c, --config string   Macaronictl configuration file
  -d, --debug           Enable debug output.
  -h, --help            help for this command
  -v, --version         version for this command

Use " [command] --help" for more information about a command.
```

## env-update

The `env-update` command follow the portage `env-update` command with
same simplification.

The generation of the `/etc/csh.env` instead is created only when
(t)csh support is enabled (with `--csh` option or through configuration
file option).

The generation of the `/etc/environment.d/10-macaroni.conf` is created only when
systemd support is enabled (with `--systemd` option or through configuration
file option).

```bash
$> macaronictl env-update

$> macaronictl env-update --dry-run

$> macaronictl env-update --csh
```

## etc-update

The `etc-update` command follows the Portage `etc-update` logic with
some simplification.

It read the same `/etc/etc-update.conf` configuration file and it permits to
use `vim`, `emacs`.

```bash
$> macaronictl etc-update
```

Could be used to analyze a specific path too:

```bash
$> macaronictl etc-update -p /opt/myconf
```

## Kernel subcommands

### Generate Initrd

```
$> macaronictl kernel gi --help
Rebuild Dracut initrd images.

$> # Generate all initrd images of the kernels available on boot dir.
$> macaronictl kernel geninitrd --all

$> # Generate all initrd images of the kernels available on boot dir
$> # and set the bzImage, Initrd links to one of the kernel available
$> # if not present or to the next release of the same kernel after the
$> # upgrade.
$> macaronictl kernel geninitrd --all --set-links

$> # Generate all initrd images of the kernels available on boot dir
$> # and set the bzImage, Initrd links to one of the kernel available
$> # if not present or to the next release of the same kernel after the
$> # upgrade. In addition, it purges old initrd images and update grub.cfg.
$> macaronictl kernel geninitrd --all --set-links --purge --grub

$> # Just show what dracut commands will be executed for every initrd images.
$> macaronictl kernel geninitrd --all --dry-run

$> # Generate the initrd image for the kernel 5.10.42
$> macaronictl kernel geninitrd --version 5.10.42

$> # Generate the initrd image for the kernel 5.10.42 and kernel type vanilla.
$> macaronictl kernel geninitrd --version 5.10.42 --ktype vanilla

$> # Generate the initrd image for the kernel 5.10.42 and kernel type vanilla
$> # and set the links bzImage, Initrd to the selected kernel/initrd.
$> macaronictl kernel geninitrd --version 5.10.42 --ktype vanilla

Usage:
   kernel geninitrd [flags]

Aliases:
  geninitrd, gi

Flags:
      --all                          Rebuild all images with kernel.
      --bootdir string               Directory where analyze kernel files. (default "/boot")
      --dracut-opts string           Override the default dracut options used on the initrd image generation.
                                     Set the MACARONICTL_DRACUT_ARGS env in alternative.
      --dry-run                      Dry run commands.
      --grub                         Update grub.cfg.
  -h, --help                         help for geninitrd
      --kernel-profiles-dir string   Specify the directory where read the kernel types profiles supported. (default "/etc/macaroni/kernels-profiles/")
      --ktype string                 Specify the kernel type of the initrd image to build.
      --purge                        Clean orphan initrd images without kernel.
      --set-links                    Set bzImage and Initrd links for the selected kernel or update links of the upgraded kernel.
      --version string               Specify the kernel version of the initrd image to build.

Global Flags:
  -c, --config string   MacaroniCtl configuration file
  -d, --debug           Enable debug output.
```

