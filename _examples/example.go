
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//     http://www.apache.org/licenses/LICENSE-2.0
// 
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
// 
// Based on Gothic examples
// Changes (c) 2015, Martin Falatic

package main

import "github.com/MartyMacGyver/gothic"
import "github.com/MartyMacGyver/gothlife"
import "image"
import "image/draw"
import "image/png"
import "image/color"
import "fmt"
import "os"
import "time"
import "math/rand"

const init_script = `
	wm title . "Conway's Life"
	grid [ttk::frame .c -padding "10 10 10 10"] -column 0 -row 0 -sticky nwes
	grid columnconfigure . 0 -weight 1; grid rowconfigure . 0 -weight 1

	ttk::button .b -text "Start" -command "proc <- 0" -state disabled -width 10
	grid .b -column 0 -row 1 -padx 2 -pady 2 -sticky nwse

	foreach w [winfo children .c] {grid configure $w -padx 5 -pady 5}
`

func loadPNG(filename string) image.Image {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}

	img, err := png.Decode(f)
	if err != nil {
		panic(err)
	}
	return img
}

func savePNG(filename string, img image.Image) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}

	err = png.Encode(f, img)
	if err != nil {
		panic(err)
	}
	return
}

func imageToRGBA(src image.Image) *image.RGBA {
	b := src.Bounds()
	m := image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
	draw.Draw(m, m.Bounds(), src, b.Min, draw.Src)
	return m
}

func proc(ir *gothic.Interpreter, labelid string, animage *image.RGBA) {
	button := ".b"
	channame := "proc"
	recv := make(chan int, 1)

	fmt.Printf("here")
	s1 := rand.NewSource(42)
    r1 := rand.New(s1)
	r1 = r1
	    
	running := false

	// register channel and enable button
	ir.RegisterCommand(channame, func(_ string, arg int) {
		select {
			case recv <- arg:
			default:
				println("! Blocked", arg)
		}
		fmt.Printf("Command received: running = %d\n", running)
	})
	ir.Eval(`%{} configure -state normal`, button)

	for {
		// wait for an event
		select {
			case <- recv:
				println("Got something", running)
				break
			default:
				println("! got nothing", running)
				time.Sleep(time.Second / 10)
				continue
		}
		//fmt.Printf("Received %d\n", arg)

		if running {
			fmt.Printf("continuing instead\n")
			continue
		}
		running = true
		fmt.Printf("Starting run\n")
		//ir.Eval(`%{} configure -state disabled -text "Running"`, button)
		ir.Eval(`%{} configure -state normal -text "Stop"`, button)
		sizex := 800
		sizey := 600
		sizex /= 2
		sizey /= 2

		animage = image.NewRGBA(image.Rect(0, 0, sizex*2, sizey*2))
		for x := 0; x < sizex; x += 1 {
			for y := 0; y < sizey; y += 1 {
				animage.Set(x,y,color.RGBA{0,0,0,255})
			}
		}

		l := gothlife.NewLife(sizex, sizey)
		for i := 0; i < 10; i++ {
			l.Step()
			for x := 0; x < sizex; x++ {
				for y := 0; y < sizey; y++ {
					state := l.CurAlive(x,y)
					thiscolor := color.RGBA{0,0,0,255}
					if state {
						thiscolor = color.RGBA{0,0,255,255}
					}
					animage.Set(x*2+0, y*2+0, thiscolor)
					animage.Set(x*2+1, y*2+0, thiscolor)
					animage.Set(x*2+0, y*2+1, thiscolor)
					animage.Set(x*2+1, y*2+1, thiscolor)
				}
			}
			t1 := time.Now()
			ir.UploadImage("bg", imageToRGBA(animage))
			fmt.Sprintf("Step %d, time %s\n", i, time.Since(t1).String())
			ir.Eval(`%{} configure -image bg`, labelid)
			time.Sleep(time.Second / 120)
		}
		// reset button state
		ir.Eval(`%{} configure -state normal -text "Start"`, button)
		running = false
		fmt.Printf("Ending run\n")
	}
}

func main() {
	ir := gothic.NewInterpreter(init_script)
    animage := image.NewRGBA(image.Rect(0, 0, 800, 600))
	labelid := ".l"
	ir.UploadImage("bg", animage)
	ir.Eval(`ttk::label %{} -image bg`, labelid)
	ir.Eval(`grid %{} -column 0 -row 5 -sticky w`, labelid)
	go proc(ir, labelid, animage)
	<-ir.Done
}
