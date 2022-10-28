echo Started
scp -o StrictHostKeyChecking=no -r -P 4242 *.yml *.go plugin go.sum go.mod ubuntu@150.136.142.44:~/devzat
echo Copied files
ssh -o StrictHostKeyChecking=no -p 4242 ubuntu@150.136.142.44 <<EOL # Unquote so lines are expanded
	cd ~/devzat
	go build && echo Built
	echo $SERVER_PASS | sudo -S pkill devzat && echo Killed
	sleep 2
	echo $SERVER_PASS | sudo -S pkill -9 devzat && echo Killed with SIGKILL
	echo $SERVER_PASS | nohup sudo -S GOMAXPROCS=2 DEVZAT_CONFIG=mainserver.yml ./devzat > /dev/null 2>stderr </dev/null &
	echo Started server
	disown
	exit
EOL
echo Finished
