pid=`ps aux | grep comake_web | grep -v grep | awk '{print $2}'`
if [ ! -z "$pid" ]; then
    kill $pid
fi
nohup ./comake_web > log.txt 2>err.log &
