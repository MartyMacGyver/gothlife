
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
	grid [ttk::frame .c -padding "10 10 10 10"] -column 0 -row 0 -sticky nwes
	grid columnconfigure . 0 -weight 1; grid rowconfigure . 0 -weight 1

	ttk::button .bctrl -text "Start" -command "mainproc 0" -state disabled -width 10
	grid .bctrl -column 0 -row 1 -padx 2 -pady 2 -sticky nwse

	ttk::button .breset -text "Reset" -command "mainproc 1" -state disabled -width 10
	grid .breset -column 0 -row 2 -padx 2 -pady 2 -sticky nwse
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

func mainproc(ir *gothic.Interpreter, labelid string, animage *image.RGBA) {
	button := ".bctrl"
	channame := "mainproc"
	recv := make(chan int, 1)
	running := false

	s1 := rand.NewSource(42)
    r1 := rand.New(s1)
	r1 = r1

	// register channel and enable button
	ir.RegisterCommand(channame, func(arg int) {
		select {
			case recv <- arg:
			default:
				println("! Blocked", arg)
		}
		fmt.Printf("Command received: running = %d\n", running)
	})
	ir.Eval(`%{} configure -state normal`, button)

	for {
		// wait for a start event
		select {
			case arg := <- recv:
				println("Got something", running, arg)
				break
			default:
				//println("Waiting", running)
				time.Sleep(time.Second / 10)
				continue
		}
		//fmt.Printf("Received %d\n", arg)

		running = true
		fmt.Printf("Starting run\n")
		ir.Eval(`%{} configure -state normal -text "Stop"`, button)
		ir.Eval(`%{} configure -state normal`, ".breset")
		sizex := 800
		sizey := 600
		sizex /= 3
		sizey /= 3

		animage = image.NewRGBA(image.Rect(0, 0, sizex*3, sizey*3))
		for x := 0; x < sizex; x += 1 {
			for y := 0; y < sizey; y += 1 {
				animage.Set(x,y,color.RGBA{0,0,0,255})
			}
		}

		life := gothlife.NewLife(sizex, sizey)
		for i := 0; running && i < 1000; i++ {
			life.Step()
			for x := 0; x < sizex; x++ {
				for y := 0; y < sizey; y++ {
					state := life.CurAlive(x,y)
					thiscolor := color.RGBA{0,0,0,255}
					if state {
						thiscolor = color.RGBA{0,0,255,255}
					}
					animage.Set(x*3+0, y*3+0, color.RGBA{0,0,0,255})
					animage.Set(x*3+1, y*3+0, thiscolor)
					animage.Set(x*3+2, y*3+0, color.RGBA{0,0,0,255})
					animage.Set(x*3+0, y*3+1, thiscolor)
					animage.Set(x*3+1, y*3+1, color.RGBA{64,0,0,255})
					animage.Set(x*3+2, y*3+1, thiscolor)
					animage.Set(x*3+0, y*3+2, color.RGBA{0,0,0,255})
					animage.Set(x*3+1, y*3+2, thiscolor)
					animage.Set(x*3+2, y*3+2, color.RGBA{0,0,0,255})
				}
			}
			t1 := time.Now()
			ir.UploadImage("bg", imageToRGBA(animage))
			fmt.Sprintf("Step %d, time %s\n", i, time.Since(t1).String())
			ir.Eval(`%{} configure -image bg`, labelid)
			time.Sleep(time.Second / 10)
			select {
				case arg := <- recv:
					println("Got an exit signal", running, arg)
					running = false
					break
				default:
					//println("Continuing", running)
			}
		}
		// reset button state
		ir.Eval(`%{} configure -state normal -text "Start"`, button)
		ir.Eval(`%{} configure -state disabled`, ".breset")
		running = false
		fmt.Printf("Ending run\n")
	}
}

func main() {
	ir := gothic.NewInterpreter(``) // init_script can go in here too
	ir.Eval(init_script)
	ir.Eval(`wm title . "%{}"`, "Conway's Life in Go")

    animage := image.NewRGBA(image.Rect(0, 0, 800, 600))
	labelid := ".l"
	ir.UploadImage("bg", animage)
	ir.Eval(`ttk::label %{} -image bg`, labelid)
	ir.Eval(`grid %{} -column 0 -row 3 -sticky w`, labelid)

	go mainproc(ir, labelid, animage)

	<-ir.Done
}
