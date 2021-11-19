# goscp

Set of tools created during my OSCP preparation. They include:

* gosheller - terminal based interface for [cmdasp.aspx](https://github.com/BlackArch/webshells/blob/master/aspx/cmdasp.aspx) webshell. It provides interactive terminal, as well as history savings.

## gosheller

```bash
$ gosheller --help
Usage of ./gosheller:
  -l	list the history
  -t int
    	timeout for server (default 5)
  -u string
    	url for the webshell
```

After you have uploaded the webshell, you can use gosheller to interact with it, all you have to do is pass the `-u` parameter to the gosheller. If your connection sucks, you can increase the timeout to something like 20(20 seconds).

```bash
$ gosheller -t 20 -u http://127.0.0.1:4167/cmdasp.aspx
> whoami
root
> hostname
lateralusd
> exit
Exiting
```

To view the history, simply pass the `-l` flag.

```bash
$ gosheller -l=true
+-------------------------------------------------------------------------------------------+
|                                      Command history                                      |
+-------------------------------+-----------------------------------+----------+------------+
| TIME                          | HOST                              | COMMAND  | OUTPUT     |
+-------------------------------+-----------------------------------+----------+------------+
| 0001-01-01 00:00:00 +0000 UTC | http://127.0.0.1:4167/cmdasp.aspx | whoami   | root       |
| 0001-01-01 00:00:00 +0000 UTC | http://127.0.0.1:4167/cmdasp.aspx | hostname | lateralusd |
+-------------------------------+-----------------------------------+----------+------------+
```
