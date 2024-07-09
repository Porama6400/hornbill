#!/bin/sh
wg genkey | tee wg-privatekey | wg pubkey > wg-publickey
ip link add dev hornbill-wg type wireguard || true
ip address add dev hornbill-wg 10.222.0.1/24 || true
ip link set up hornbill-wg || true
wg set hornbill-wg listen-port 51820 private-key wg-privatekey
