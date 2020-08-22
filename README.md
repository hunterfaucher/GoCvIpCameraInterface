# RSTP Feed on HTTP Server in Golang 

## Install OpenCV for Go
```
$ go get -u -d gocv.io/x/gocv
$ cd $GOPATH/src/gocv.io/x/gocv
$ make install
```
## Run it
```
$ go run webview.go -xml /cameraProject/files/haarcascade_frontalface_default.xml -target1 rtsp://192.168.1.88/11
