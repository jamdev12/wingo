package commands

import (
	"strings"

	"github.com/BurntSushi/gribble"

	"github.com/BurntSushi/xgb/xproto"

	"github.com/BurntSushi/wingo/focus"
	"github.com/BurntSushi/wingo/logger"
	"github.com/BurntSushi/wingo/prompt"
	"github.com/BurntSushi/wingo/workspace"
	"github.com/BurntSushi/wingo/wm"
	"github.com/BurntSushi/wingo/xclient"
)

// parsePos takes a string and parses an x or y position from it.
// The magic here is that while a string could just be a simple integer,
// it could also be a float greater than 0 but <= 1 in terms of the current
// head's geometry.
func parsePos(gribblePos gribble.Any, y bool) (int, bool) {
	switch pos := gribblePos.(type) {
	case int:
		return pos, true
	case float64:
		if pos <= 0 || pos > 1 {
			logger.Warning.Printf("'%s' not in the valid range (0, 1].", pos)
			return 0, false
		}

		headGeom := wm.Workspace().Geom()
		if y {
			return headGeom.Y() + int(float64(headGeom.Height())*pos), true
		} else {
			return headGeom.X() + int(float64(headGeom.Width())*pos), true
		}
	}
	panic("unreachable")
}

// parseDim takes a string and parses a width or height dimension from it.
// The magic here is that while a string could just be a simple integer,
// it could also be a float greater than 0 but <= 1 in terms of the current
// head's geometry.
func parseDim(gribbleDim gribble.Any, height bool) (int, bool) {
	switch dim := gribbleDim.(type) {
	case int:
		return dim, true
	case float64:
		if dim <= 0 || dim > 1 {
			logger.Warning.Printf("'%s' not in the valid range (0, 1].", dim)
			return 0, false
		}

		headGeom := wm.Workspace().Geom()
		if height {
			return int(float64(headGeom.Height()) * dim), true
		} else {
			return int(float64(headGeom.Width()) * dim), true
		}
	}
	panic("unreachable")
}

// stringBool takes a string and returns true if the string corresponds
// to a "true" value. i.e., "Yes", "Y", "y", "YES", "yEs", etc.
func stringBool(s string) bool {
	sl := strings.ToLower(s)
	return sl == "yes" || sl == "y"
}

// stringTabComp takes a string and converts it to a tab completion constant
// defined in the prompt package. Valid values are "Prefix" and "Any."
func stringTabComp(s string) int {
	switch s {
	case "Prefix":
		return prompt.TabCompletePrefix
	case "Any":
		return prompt.TabCompleteAny
	default:
		logger.Warning.Printf(
			"Tab completion mode '%s' not supported.", s)
	}
	return prompt.TabCompletePrefix
}

// Shortcut for executing Client interface functions that have no parameters
// and no return values on the currently focused window.
func withFocused(f func(c *xclient.Client)) gribble.Any {
	if focused := focus.Current(); focused != nil {
		client := focused.(*xclient.Client)
		f(client)
		return int(client.Id())
	}
	return ":void:"
}

func withClient(cArg gribble.Any, f func(c *xclient.Client)) gribble.Any {
	switch c := cArg.(type) {
	case int:
		if c == 0 {
			return withFocused(f)
		}
		for _, client_ := range wm.Clients {
			client := client_.(*xclient.Client)
			if int(client.Id()) == c {
				f(client)
				return int(client.Id())
			}
		}
		return ":void:"
	case string:
		switch c {
		case ":void:":
			return ":void:"
		case ":mouse:":
			wid := xproto.Window(wm.MouseClientClicked)
			if client := wm.FindManagedClient(wid); client != nil {
				c := client.(*xclient.Client)
				f(c)
				return int(c.Id())
			} else {
				f(nil)
				return ":void:"
			}
		case ":active:":
			return withFocused(f)
		default:
			panic("Client name Not implemented: " + c)
		}
	}
	panic("unreachable")
}

func withWorkspace(wArg gribble.Any, f func(wrk *workspace.Workspace)) {
	switch w := wArg.(type) {
	case int:
		if wrk := wm.Heads.Workspaces.Get(w); wrk != nil {
			f(wrk)
		}
	case string:
		if wrk := wm.Heads.Workspaces.Find(w); wrk != nil {
			f(wrk)
		}
	}
}