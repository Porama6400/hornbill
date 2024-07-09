#!/bin/sh
cp bin/* /usr/local/bin

mkdir /etc/hornbill > /dev/null 2> /dev/null || true
cp etc/* /etc/hornbill

cp systemd/hornbill.service /etc/systemd/system
