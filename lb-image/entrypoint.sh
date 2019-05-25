#!/bin/sh
set -e -u

trap exit TERM INT

# Validate that we can operate
if [ `cat /proc/sys/net/ipv4/ip_forward` != 1 ]; then
    echo "ip_forward is not enabled"
    exit 1
fi

echo "Setting up forwarding from port $SRC_PORT to $DEST_IP:$DEST_PORT/$DEST_PROTO"

# Setup the actual forwarding
sudo iptables -t nat -I PREROUTING ! -s ${DEST_IP}/32 -p ${DEST_PROTO} --dport ${SRC_PORT} -j DNAT --to ${DEST_IP}:${DEST_PORT}
sudo iptables -t nat -I POSTROUTING -d ${DEST_IP}/32 -p ${DEST_PROTO} -j MASQUERADE


echo "Forwarding set up succesfully, taking Cinderella nap..."

if [ ! -e /tmp/pause ]; then
    mkfifo /tmp/pause
fi
</tmp/pause &
wait $!
