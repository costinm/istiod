
# Root install

- Shell script creates a systemd unit, start by default
- systemctl stop k3s, disable k3s

- kubeconfig:  /etc/rancher/k3s/k3s.yaml
- /etc/systemd/system/k3s.service
- /etc/systemd/system/k3s.service.env

```shell script

cat /etc/systemd/system/k3s.service.env 
K3S_KUBECONFIG_MODE=666


journalctl -u k3s

```

/var/lib/rancher
