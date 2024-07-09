package main

import (
	"fmt"
	"sort"

	"github.com/jroimartin/gocui"
)

type KeyBind struct {
	viewname string
	key      interface{}
	mod      gocui.Modifier
	handler  func(*gocui.Gui, *gocui.View) error
}

func (app *App) bindings() error {
	bindings := []KeyBind{
		{"", gocui.KeyCtrlC, gocui.ModNone, app.quit},
		{"", 'w', gocui.ModNone, app.up},
		{"", 's', gocui.ModNone, app.down},
		{"", 'p', gocui.ModNone, app.photo},
	}

	for _, b := range bindings {
		if err := app.g.SetKeybinding(b.viewname, b.key, b.mod, b.handler); err != nil {
			return err
		}
	}

	return nil
}

func (app *App) redraw() {
	app.g.Update(func(gui *gocui.Gui) error {
		if v, err := gui.View("info"); err == nil {
			v.Clear()
			fmt.Fprintf(v, "state: %s\n", app.info.getString("state"))
			sat := fmt.Sprintf("%d", app.info.getByte("sat"))

			if app.info.getBool("sat_good") {
				sat = WithColors(sat, FgGreen)
			} else {
				sat = WithColors(sat, FgRed, Bold)
			}
			fmt.Fprintf(v, "sat: %s\n", sat)

			fmt.Fprintf(v, "lat: %.7f lon: %.7f\n", app.info.getFloat("lat"), app.info.getFloat("lon"))
			fmt.Fprintf(v, "roll: %s pitch: %s yaw: %7.2f\n",
				formatGyro(app.info.getFloat("roll")),
				formatGyro(app.info.getFloat("pitch")),
				app.info.getFloat("yaw"))

			fmt.Fprintf(v, "alt: %.1f dist %.0f\n", app.info.getFloat("alt"), app.info.getFloat("dist"))
			fmt.Fprintf(v, "em: %d battery: %.2fv\n", app.info.getByte("em"), app.info.getFloat("battery"))

			fmt.Fprintf(v, "\n\n")
			m := make(map[string]any)
			keys := make([]string, 0)
			app.info.info.Range(func(key, value any) bool {
				keys = append(keys, key.(string))
				m[key.(string)] = value
				return true
			})
			sort.Strings(keys)
			i := 0
			for _, k := range keys {
				fmt.Fprintf(v, "%s: %v ", k, m[k])
				i++
				if i%4 == 0 {
					fmt.Fprintf(v, "\n")
				}
			}
		}

		if v, err := gui.View("data"); err == nil {
			v.Clear()
			m := make(map[int]string)
			keys := make([]int, 0)
			app.data.Range(func(key, value any) bool {
				m[int(key.(byte))] = arr2str(value.([]byte))
				keys = append(keys, int(key.(byte)))
				return true
			})
			sort.Ints(keys)
			for _, k := range keys {
				fmt.Fprintf(v, "%.2x: %s\n", k, m[k])
			}
		}
		if v, err := gui.View("log"); err == nil {
			v.Clear()
			_, size := v.Size()
			for _, l := range app.logger.GetLines(size) {
				fmt.Fprintln(v, l)
			}
		}

		return nil
	})
}

func formatGyro(v float64) string {
	s := fmt.Sprintf("%7.2f", v)
	if v > -5 && v < 5 {
		return WithColors(s, FgGreen)
	}

	if v > -30 && v < 30 {
		return WithColors(s, FgYellow)
	}

	return WithColors(s, Bold, FgRed)
}
