package i3

import (
	"github.com/BurntSushi/xgbutil"
	"github.com/BurntSushi/xgbutil/xprop"
)

// IsRunningHook provides a method to override the method which detects if i3 is running or not
var IsRunningHook = func() bool {
	xu, err := xgbutil.NewConn()
	if err != nil {
		return false // X session terminated
	}
	defer xu.Conn().Close()
	reply, err := xprop.GetProperty(xu, xu.RootWin(), "I3_PID")
	if err != nil {
		return false // I3_PID no longer present (X session replaced?)
	}
	num, err := xprop.PropValNum(reply, err)
	if err != nil {
		return false
	}
	return pidValid(int(num))
}

func i3Running() bool {
	return IsRunningHook()
}
