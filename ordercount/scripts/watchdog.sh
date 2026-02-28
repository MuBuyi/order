#!/bin/bash
# 简单健康检查守护脚本：每 30s ping 应用首页，若连续 2 次失败则重启 app 服务并记录日志
LOGDIR="/home/condingyang/coding/order/ordercount/logs"
mkdir -p "$LOGDIR"
LOGFILE="$LOGDIR/watchdog.log"
FAIL_COUNT=0
THRESHOLD=2
while true; do
  HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" --max-time 5 http://127.0.0.1:8080/ || echo "000")
  TIMESTAMP=$(date +"%F %T")
  if [ "$HTTP_CODE" != "200" ]; then
    FAIL_COUNT=$((FAIL_COUNT+1))
    echo "$TIMESTAMP - unhealthy (code=$HTTP_CODE), fail_count=$FAIL_COUNT" >> "$LOGFILE"
  else
    if [ $FAIL_COUNT -ne 0 ]; then
      echo "$TIMESTAMP - recovered (code=$HTTP_CODE)" >> "$LOGFILE"
    fi
    FAIL_COUNT=0
  fi
  if [ $FAIL_COUNT -ge $THRESHOLD ]; then
    echo "$TIMESTAMP - restarting app because fail_count=$FAIL_COUNT" >> "$LOGFILE"
    sudo docker compose restart app >> "$LOGFILE" 2>&1 || sudo docker restart ordercount-app >> "$LOGFILE" 2>&1
    FAIL_COUNT=0
  fi
  sleep 30
done
