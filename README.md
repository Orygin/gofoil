# Gofoil
Gofoil is a little tool inspired by tinfoil's python [remote_install_pc](https://github.com/Adubbz/Tinfoil/blob/master/tools/remote_install_pc.py), written in go. Using a little bit of [bycEEE/tinfoilusbgo](github.com/bycEEE/tinfoilusbgo).  

## Why
I wanted to load switch saves from my NAS onto my switch without powering up my computer.
No client provided samba or NFS, so this was designed as an always on server to run on my NAS, that can handshake with Tinfoil-like net install
and serve files directly from there.  

This was tested only with Awoo Installer on Atmosphere. May also work on tinfoil but not tested.
 
## Usage
To run gofoil, you need to provide it a few arguments, like the host IP, port, and folder to scan for switch files.  
Example, on a windows machine (and shared network):   
```shell script
gofoil.exe -root Z:\\ -folders Downloads,Games/switch -ip 192.168.1.95 -port 8000
```
It will start a server at the ip:port indicated. You can open this page (on your phone for example) to show a page where you can input your switch's IP.  
With Awoo or Tinfoil opened on your switch, and the network install selected, you should now see a list of files from the directory, available to install.

## Building
Use the regular go command to get the project:
```shell script
go get -u gofoil
```
You can build for another OS/Arch target using the env vars `GOOS` and `GOARCH`  
For example for my NAS I had to build for linux/armv5:
```shell script
GOOS="linux" GOARCH="arm" GOARM=5 go build
```
