
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

const appTitle  = "Conway's Life in Go"
const bControl  = ".bctrl"
const bReset    = ".breset"
const lifeField = ".lfield"
const lifeImage = ".limage"
const mainchan  = "controlchannel"

const init_script = `
	grid [ttk::frame .c -padding "10 10 10 10"] -column 0 -row 0 -sticky nwes
	grid columnconfigure . 0 -weight 1; grid rowconfigure . 0 -weight 1
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

func mainproc(ir *gothic.Interpreter, animage *image.RGBA, gridx int, gridy int) {
	recv := make(chan int, 1)
	running := false
	lifeMaxGens := 50

	s1 := rand.NewSource(42)
    r1 := rand.New(s1)
	r1 = r1

	ir.RegisterCommand(mainchan, func(arg int) {
		select {
			case recv <- arg:
			default:
				println("! Blocked", arg)
		}
		fmt.Printf("Command received: running = %d\n", running)
	})
	ir.Eval(`%{} configure -state normal`, bControl)

	for {
		select {
			case arg := <- recv:
				println("Got start", running, arg)
				break
			default:
				time.Sleep(time.Second / 10)
				continue
		}

		running = true
		fmt.Printf("Starting run\n")
		ir.Eval(`%{} configure -state normal -text "Stop"`, bControl)
		ir.Eval(`%{} configure -state normal`, bReset)
		scale := 3
		sizex := gridx/scale
		sizey := gridy/scale

		animage = image.NewRGBA(image.Rect(0, 0, sizex*scale, sizey*scale))
		for x := 0; x < sizex*scale; x += 1 {
			for y := 0; y < sizey*scale; y += 1 {
				animage.Set(x,y,color.RGBA{0,0,0,255})
			}
		}

		life := gothlife.NewLife(sizex, sizey)
		for i := 0; running && i < lifeMaxGens; i++ {
			life.Step()
			for x := 0; x < sizex; x++ {
				for y := 0; y < sizey; y++ {
					state := life.CurAlive(x,y)
					thiscolor := color.RGBA{0,0,0,255}
					if state {
						thiscolor = color.RGBA{0,0,255,255}
					}
					animage.Set(x*scale+0, y*scale+0, color.RGBA{0,0,0,255})
					animage.Set(x*scale+1, y*scale+0, thiscolor)
					animage.Set(x*scale+2, y*scale+0, color.RGBA{0,0,0,255})
					animage.Set(x*scale+0, y*scale+1, thiscolor)
					animage.Set(x*scale+1, y*scale+1, color.RGBA{64,0,0,255})
					animage.Set(x*scale+2, y*scale+1, thiscolor)
					animage.Set(x*scale+0, y*scale+2, color.RGBA{0,0,0,255})
					animage.Set(x*scale+1, y*scale+2, thiscolor)
					animage.Set(x*scale+2, y*scale+2, color.RGBA{0,0,0,255})
				}
			}
			t1 := time.Now()
			ir.UploadImage(lifeImage, imageToRGBA(animage))
			fmt.Sprintf("Step %d, time %s\n", i, time.Since(t1).String())
			ir.Eval(`%{} configure -image %{}`, lifeField, lifeImage)
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

		ir.Eval(`%{} configure -state normal -text "Start"`, bControl)
		ir.Eval(`%{} configure -state disabled`, bReset)
		running = false
		fmt.Printf("Ending run\n")
	}
}

func main() {
	ir := gothic.NewInterpreter(``) // init_script can go in here too
	ir.Eval(init_script)
	ir.Eval(`wm title . "%{}"`, appTitle)
	gridx := 320
	gridy := 240

	row := 0
	col := 0
	ir.Eval(`ttk::button %{} -text "Start" -command "%{} 0" -state disabled -width 10`, bControl, mainchan)
	ir.Eval(`grid %{} -column %{} -row %{} -padx 2 -pady 2 -sticky nwse`, bControl, col, row)

	row += 1
	col = 0
	ir.Eval(`ttk::button %{} -text "Reset" -command "%{} 1" -state disabled -width 10`, bReset, mainchan)
	ir.Eval(`grid %{} -column %{} -row %{} -padx 2 -pady 2 -sticky nwse`, bReset, col, row)

	row += 1
	col = 0
    animage := image.NewRGBA(image.Rect(0, 0, gridx, gridy))
	ir.UploadImage(lifeImage, animage)
	ir.Eval(`ttk::label %{} -image %{}`, lifeField, lifeImage)
	ir.Eval(`grid %{} -column %{} -row %{} -sticky w`, lifeField, col, row)

	go mainproc(ir, animage, gridx, gridy)

	<-ir.Done
}
