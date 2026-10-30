package signal

import (
	"syscall"
)

// SignalMap is a map of Haiku signals (air/OS: <signal.h>'s numbering).
var SignalMap = map[string]syscall.Signal{
	"HUP":    syscall.Signal(1),
	"INT":    syscall.Signal(2),
	"QUIT":   syscall.Signal(3),
	"ILL":    syscall.Signal(4),
	"CHLD":   syscall.Signal(5),
	"ABRT":   syscall.Signal(6),
	"IOT":    syscall.Signal(6),
	"PIPE":   syscall.Signal(7),
	"FPE":    syscall.Signal(8),
	"KILL":   syscall.Signal(9),
	"STOP":   syscall.Signal(10),
	"SEGV":   syscall.Signal(11),
	"CONT":   syscall.Signal(12),
	"TSTP":   syscall.Signal(13),
	"ALRM":   syscall.Signal(14),
	"TERM":   syscall.Signal(15),
	"TTIN":   syscall.Signal(16),
	"TTOU":   syscall.Signal(17),
	"USR1":   syscall.Signal(18),
	"USR2":   syscall.Signal(19),
	"WINCH":  syscall.Signal(20),
	"TRAP":   syscall.Signal(22),
	"POLL":   syscall.Signal(23),
	"IO":     syscall.Signal(23),
	"PROF":   syscall.Signal(24),
	"SYS":    syscall.Signal(25),
	"URG":    syscall.Signal(26),
	"VTALRM": syscall.Signal(27),
	"XCPU":   syscall.Signal(28),
	"XFSZ":   syscall.Signal(29),
	"BUS":    syscall.Signal(30),
}
