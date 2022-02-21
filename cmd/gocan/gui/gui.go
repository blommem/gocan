package gui

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/data/binding"
	flayout "fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/widget"
	"github.com/roffe/gocan/pkg/ecu"
	"github.com/roffe/gocan/pkg/t5"
	"github.com/roffe/gocan/pkg/t7"
	"github.com/roffe/gocan/pkg/t8"
	sdialog "github.com/sqweek/dialog"
	"go.bug.st/serial/enumerator"
)

type mainWindow struct {
	app    fyne.App
	window fyne.Window

	log      *widget.List
	infoBTN  *widget.Button
	dumpBTN  *widget.Button
	flashBTN *widget.Button

	ecuList     *widget.Select
	adapterList *widget.Select
	portList    *widget.Select
	speedList   *widget.Select

	progressBar *widget.ProgressBar
}

type appState struct {
	ecuType   ecu.Type
	canRate   float64
	adapter   string
	port      string
	portSpeed int
}

var (
	mw      *mainWindow
	logData []string
	state   *appState
)

func init() {
	state = &appState{}
}

func Run(ctx context.Context) {
	a := app.New()
	a.Settings().SetTheme(&gocanTheme{})

	w := a.NewWindow("GoCANFlasher")
	w.Resize(fyne.NewSize(900, 500))

	mw = &mainWindow{
		app:    a,
		window: w,
		log: widget.NewList(
			func() int {
				return len(logData)
			},
			func() fyne.CanvasObject {
				return widget.NewLabel("template")
			},
			func(i widget.ListItemID, o fyne.CanvasObject) {
				o.(*widget.Label).SetText(logData[i])
			},
		),

		infoBTN:  widget.NewButton("Info", ecuInfo),
		dumpBTN:  widget.NewButton("Dump", ecuDump),
		flashBTN: widget.NewButton("Flash", ecuFlash),
	}

	mw.ecuList = widget.NewSelect([]string{"Trionic 5", "Trionic 7", "Trionic 8"}, func(s string) {
		state.ecuType = ecu.Type(mw.ecuList.SelectedIndex() + 1)
		log.Println(state.ecuType.String())
		switch state.ecuType {
		case ecu.Trionic5:
			state.canRate = t5.PBusRate
		case ecu.Trionic7:
			state.canRate = t7.PBusRate
		case ecu.Trionic8:
			state.canRate = t8.PBusRate
		}

	})
	mw.adapterList = widget.NewSelect(adapters(), func(s string) {
		state.adapter = s
	})
	mw.portList = widget.NewSelect(ports(), func(s string) {
		state.port = s
	})
	mw.speedList = widget.NewSelect(speeds(), func(s string) {
		speed, err := strconv.Atoi(s)
		if err != nil {
			output("failed to set port speed: " + err.Error())
		}
		state.portSpeed = speed
	})

	mw.ecuList.PlaceHolder = "Select ECU"
	mw.adapterList.PlaceHolder = "Select Adapter"
	mw.portList.PlaceHolder = "Select Port"
	mw.speedList.PlaceHolder = "Select Speed"

	progress := binding.NewFloat()
	mw.progressBar = widget.NewProgressBarWithData(progress)
	mw.progressBar.Max = 100
	mw.progressBar.Hide()

	left := container.New(flayout.NewMaxLayout(), mw.log)
	widget.NewToolbar()
	right := container.NewVBox(
		mw.infoBTN,
		mw.dumpBTN,
		mw.flashBTN,
		widget.NewSeparator(),
		mw.ecuList,
		mw.adapterList,
		mw.portList,
		mw.speedList,
		widget.NewSeparator(),
		mw.progressBar,
	)

	split := container.NewHSplit(left, right)
	split.Offset = 0.7
	w.SetContent(split)

	go func() {
		<-ctx.Done()
		w.Close()
	}()

	w.ShowAndRun()
}

func checkSelections() bool {
	var out strings.Builder

	if mw.ecuList.SelectedIndex() < 0 {
		out.WriteString("ECU type\n")
	}

	if mw.adapterList.SelectedIndex() < 0 {
		out.WriteString("Adapter\n")
	}
	if mw.portList.SelectedIndex() < 0 {
		out.WriteString("Port\n")
	}
	if mw.speedList.SelectedIndex() < 0 {
		out.WriteString("Speed\n")
	}
	if out.Len() > 0 {
		sdialog.Message("Please set the following options:\n%s", out.String()).
			Title("error").
			Error()
		return false
	}
	return true
}

func output(s string) {
	text := "\n"
	if s != "" {
		text = fmt.Sprintf("%s - %s\n", time.Now().Format("15:04:05.000"), s)
	}
	logData = append(logData, text)
	mw.log.Refresh()
	mw.log.ScrollToBottom()
}

func adapters() []string {
	return []string{"CanUSB", "OBDLinkSX"}
}

func speeds() []string {
	var out []string
	l := []int{9600, 19200, 38400, 57600, 115200, 230400, 460800, 921600, 1000000, 2000000}
	for _, ll := range l {
		out = append(out, strconv.Itoa(ll))
	}
	return out
}

func disableButtons() {
	mw.infoBTN.Disable()
	mw.dumpBTN.Disable()
	mw.flashBTN.Disable()
}

func enableButtons() {
	mw.infoBTN.Enable()
	mw.dumpBTN.Enable()
	mw.flashBTN.Enable()
}

func ports() []string {
	var portsList []string
	ports, err := enumerator.GetDetailedPortsList()
	if err != nil {
		log.Fatal(err)
	}
	if len(ports) == 0 {
		output("No serial ports found!")
		return nil
	}
	for _, port := range ports {
		output(fmt.Sprintf("Found port: %s", port.Name))
		if port.IsUSB {
			output(fmt.Sprintf("  USB ID     %s:%s", port.VID, port.PID))
			output(fmt.Sprintf("  USB serial %s", port.SerialNumber))
			output("")
			portsList = append(portsList, port.Name)
		}
	}
	return portsList
}
