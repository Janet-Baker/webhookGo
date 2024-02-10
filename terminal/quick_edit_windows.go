//go:build windows

package terminal

import (
	"golang.org/x/sys/windows"
	"os"
)

// DisableQuickEdit 禁用快速编辑
func DisableQuickEdit() error {
	stdin := windows.Handle(os.Stdin.Fd())

	var mode uint32
	err := windows.GetConsoleMode(stdin, &mode)
	if err != nil {
		return err
	}

	// 禁用快速编辑模式
	// 禁用鼠标输入
	// 禁用插入模式
	// 禁用窗口输入
	// 禁用虚拟终端输入
	mode &^= windows.ENABLE_QUICK_EDIT_MODE |
		windows.ENABLE_MOUSE_INPUT |
		windows.ENABLE_INSERT_MODE |
		windows.ENABLE_WINDOW_INPUT |
		windows.ENABLE_VIRTUAL_TERMINAL_INPUT
	// 启用扩展标志
	// 启用控制输入
	// 启用输入回显
	// 启用逐行输入
	mode |= windows.ENABLE_EXTENDED_FLAGS |
		windows.ENABLE_PROCESSED_INPUT |
		windows.ENABLE_ECHO_INPUT |
		windows.ENABLE_LINE_INPUT

	return windows.SetConsoleMode(stdin, mode)
}
