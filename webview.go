package main

import (
	"flag"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"log"
	"net/http"
	"sync"
	"time"

	"gocv.io/x/gocv"
)

var (
	err     error
	webcam  *gocv.VideoCapture
	frameid int
)

var buffer = make(map[int][]byte)
var frame []byte
var mutex = &sync.Mutex{}
var mutex2 = &sync.Mutex{}
var classifierFile string
var imgcv = gocv.NewMat()

func main() {

	host := "0.0.0.0:3000"

	target1 := flag.String("target1", "", "enter camera id or rtsp ip")

	xmlFile := flag.String("xml", "", "path to xml file")
	flag.Parse()
	if *target1 == "" || *xmlFile == "" {
		flag.Usage()
		return
	}

	defer imgcv.Close()

	classifierFile = *xmlFile

	webcam, err = gocv.OpenVideoCapture(*target1)

	if err != nil {
		fmt.Printf("Error opening capture device: \n")
		return
	}
	defer webcam.Close()

	// start capturing
	go getframes()
	go processFrame()

	fmt.Println("Capturing. Open http://" + host)

	// start http server
	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		data := ""
		for {
			//fmt.Println("Frame ID: ", frameid)
			mutex.Lock()
			data = "--frame\r\n  Content-Type: image/jpeg\r\n\r\n" + string(frame) + "\r\n\r\n"
			mutex.Unlock()
			time.Sleep(33 * time.Millisecond)
			w.Write([]byte(data))
		}
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		t, _ := template.ParseFiles("index.html")
		t.Execute(w, "index")
	})

	log.Fatal(http.ListenAndServe(host, nil))
}

func processFrame() {
	blue := color.RGBA{0, 0, 255, 0}

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(classifierFile) {
		fmt.Printf("Error reading cascade file: %v\n", classifierFile)
		return
	}
	for {

		if imgcv.Empty() {
			continue
		}

		mutex2.Lock()
		rects := classifier.DetectMultiScale(imgcv)

		fmt.Printf("found %d faces\n", len(rects))
		//i := 0
		//mutex.Lock()

		for _, r := range rects {
			//fmt.Printf("%v\n", r)
			//i++
			//mutex.Lock()
			gocv.Rectangle(&imgcv, r, blue, 3)
			//mutex.UnLock()
			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&imgcv, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)
		}
		mutex2.Unlock()
	}

}

func getframes() {

	img := gocv.NewMat()
	defer img.Close()
	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed\n")
			return
		}
		if img.Empty() {
			continue
		}
		frameid++

		imgcv = img

		gocv.Resize(img, &img, image.Point{}, float64(0.5), float64(0.5), 0)
		frame, _ = gocv.IMEncode(".jpg", img)

	}
	/*
		for {
			if ok := webcam.Read(&img); !ok {
				fmt.Printf("Device closed\n")
				return
			}
			if img.Empty() {
				continue
			}

			img2 := img

			frame, _ = gocv.IMEncode(".jpg", img)

			gocv.Resize(img2, &img2, image.Point{}, float64(0.5), float64(0.5), 0)

			rects := classifier.DetectMultiScale(img2)

			fmt.Printf("found %d faces\n", len(rects))
			//i := 0
			//mutex.Lock()

			for _, r := range rects {
				//fmt.Printf("%v\n", r)
				//i++
				//mutex.Lock()
				gocv.Rectangle(&img2, r, blue, 3)
				//mutex.UnLock()
				size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
				pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
				gocv.PutText(&img2, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)

			}

			time.Sleep(1 * time.Millisecond)
			//frameid++
			//fmt.Printf("%v\n", frameid)

			//frame, _ = gocv.IMEncode(".jpg", img)

		}*/
}
