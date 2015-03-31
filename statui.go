package main

import "fmt"
import "time"
import "log"
import "os"
import "os/exec"
import "strconv"
import "strings"
import ui "github.com/gizak/termui"
import tm "github.com/nsf/termbox-go"

type Stat struct {
	Pid, Virt, Res int
	Cpu, Mem       float64
	Command        string
}

func get_stat_of_pid(pid int) (s Stat) {
	cmd := fmt.Sprintf("top -b -n1 -p %v", pid)
	o, err := exec.Command("sh", "-c", cmd).Output()
	if err != nil {
		log.Fatal(err)
	}
	t := strings.Fields(strings.Split(string(o), "\n")[7])
	// fmt.Println(t)
	s.Pid, _ = strconv.Atoi(t[0])
	s.Virt, _ = strconv.Atoi(t[4])
	s.Res, _ = strconv.Atoi(t[5])
	s.Cpu, _ = strconv.ParseFloat(t[8], 64)
	s.Mem, _ = strconv.ParseFloat(t[9], 64)
	s.Command = t[11]
	return
}

func main() {
	// check usage
	args := os.Args[1:]
	if len(args) != 1 {
		fmt.Println(" Usage: cstat <PID> ")
		return
	}

	// init ui
	err := ui.Init()
	if err != nil {
		panic(err)
	}
	defer ui.Close()

	// set theme
	ui.UseTheme("helloworld")

	// add header
	header := ui.NewPar("process monitoring dashboard")
	header.Height = 5
	header.Border.Label = "Header"
	header.PaddingTop = 1

	// add mem% line chart
	lc_mem := ui.NewLineChart()
	lc_mem.Border.Label = "Mem%"
	lc_mem.Data = nil
	lc_mem.Height = 20
	lc_mem.AxesColor = ui.ColorWhite
	lc_mem.LineColor = ui.ColorRed | ui.AttrBold
	lc_mem.PaddingTop = 1

	// add cpu% line chart
	lc_cpu := ui.NewLineChart()
	lc_cpu.Border.Label = "Cpu%"
	lc_cpu.Data = nil
	lc_cpu.Height = 20
	lc_cpu.AxesColor = ui.ColorWhite
	lc_cpu.LineColor = ui.ColorYellow | ui.AttrBold
	lc_cpu.PaddingTop = 1

	// build layout
	ui.Body.AddRows(
		ui.NewRow(
			ui.NewCol(12, 0, header)),
		ui.NewRow(
			ui.NewCol(6, 0, lc_mem),
			ui.NewCol(6, 0, lc_cpu)))

	// calculate layout
	ui.Body.Align()

	draw := func() {
		args := os.Args[1:]
		pid, _ := strconv.Atoi(args[0])
		s := get_stat_of_pid(pid)
		lc_mem.Data = append(lc_mem.Data, s.Mem)
		lc_cpu.Data = append(lc_cpu.Data, s.Cpu)
		if len(lc_cpu.Data) > 200 {
			lc_mem.Data = lc_mem.Data[1:]
			lc_cpu.Data = lc_cpu.Data[1:]
		}
		ui.Render(ui.Body)
		// fmt.Println(s)
	}

	evt := make(chan tm.Event)

	go func() {
		for {
			evt <- tm.PollEvent()
		}
	}()

	for {
		select {
		case e := <-evt:
			if e.Type == tm.EventKey && e.Ch == 'q' {
				return
			}
			if e.Type == tm.EventResize {
				ui.Body.Width = ui.TermWidth()
				ui.Body.Align()
			}
		default:
			draw()
			time.Sleep(time.Second / 10)
		}
	}
}
