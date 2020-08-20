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
	webcam1 *gocv.VideoCapture
	frameid int
)

var buffer = make(map[int][]byte)
var frame []byte
var mutex = &sync.Mutex{}
var classifierFile string

func main() {

	host := "0.0.0.0:3000"

	target1 := flag.String("target1", "", "enter camera id or rtsp ip")
	target2 := flag.String("target2", "", "enter camera id or rtsp ip")
	xmlFile := flag.String("xml", "", "path to xml file")
	flag.Parse()
	if *target1 == "" || *xmlFile == "" || *target2 == "" {
		flag.Usage()
		return
	}

	classifierFile = *xmlFile

	webcam, err = gocv.OpenVideoCapture(*target1)

	if err != nil {
		fmt.Printf("Error opening capture device: \n")
		return
	}
	defer webcam.Close()

	webcam1, err = gocv.VideoCaptureFile(*target2)

	if err != nil {
		fmt.Printf("Error opening capture device: \n")
		return
	}
	defer webcam1.Close()

	// start capturing
	go getframes()

	fmt.Println("Capturing. Open http://" + host)

	// start http server
	http.HandleFunc("/video", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
		data := ""
		for {
			/*			fmt.Println("Frame ID: ", frame_id)
			 */mutex.Lock()
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

func getframes() {
	img := gocv.NewMat()
	defer img.Close()

	gimg := gocv.NewMat()
	defer gimg.Close()

	blue := color.RGBA{0, 0, 255, 0}

	classifier := gocv.NewCascadeClassifier()
	defer classifier.Close()

	if !classifier.Load(classifierFile) {
		fmt.Printf("Error reading cascade file: %v\n", classifierFile)
		return
	}

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed\n")
			return
		}
		if img.Empty() {
			continue
		}

		if ok := webcam1.Read(&gimg); !ok {
			fmt.Printf("Device closed\n")
			return
		}
		if gimg.Empty() {
			continue
		}

		gocv.Resize(img, &img, image.Point{}, float64(0.5), float64(0.5), 0)
		gocv.Resize(gimg, &gimg, image.Point{}, float64(0.5), float64(0.5), 0)

		rects := classifier.DetectMultiScale(gimg)

		//fmt.Printf("found %d faces\n", len(rects))
		//i := 0
		//mutex.Lock()
		for _, r := range rects {
			//fmt.Printf("%v\n", r)
			//i++
			//mutex.Lock()
			gocv.Rectangle(&img, r, blue, 3)
			//mutex.UnLock()
			size := gocv.GetTextSize("Human", gocv.FontHersheyPlain, 1.2, 2)
			pt := image.Pt(r.Min.X+(r.Min.X/2)-(size.X/2), r.Min.Y-2)
			gocv.PutText(&img, "Human", pt, gocv.FontHersheyPlain, 1.2, blue, 2)

		}

		//time.Sleep(1 * time.Millisecond)
		//frameid++
		//fmt.Printf("%v\n", frameid)

		frame, _ = gocv.IMEncode(".jpg", img)

	}
}
