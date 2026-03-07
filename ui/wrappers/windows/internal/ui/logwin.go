package ui

import (
	"log"
	"os/exec"
	"strings"

	"timepad/windows/internal/logger"
)

func ShowLogs(buf *logger.Buffer) {
	lines := buf.Lines()
	if len(lines) == 0 {
		lines = []string{"no logs yet"}
	}

	// Escape single-quotes for PowerShell heredoc-style string.
	content := strings.Join(lines, "\n")
	content = strings.ReplaceAll(content, "'", "\u2019")

	script := `
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing

$form = New-Object System.Windows.Forms.Form
$form.Text    = 'Timepad — Logs'
$form.Size    = New-Object System.Drawing.Size(960, 640)
$form.StartPosition = 'CenterScreen'
$form.BackColor     = [System.Drawing.Color]::FromArgb(18, 18, 18)

$panel = New-Object System.Windows.Forms.Panel
$panel.Dock = 'Top'
$panel.Height = 36
$panel.BackColor = [System.Drawing.Color]::FromArgb(28, 28, 28)

$btnCopy = New-Object System.Windows.Forms.Button
$btnCopy.Text = 'Copy All'
$btnCopy.FlatStyle = 'Flat'
$btnCopy.FlatAppearance.BorderSize = 0
$btnCopy.ForeColor = [System.Drawing.Color]::White
$btnCopy.BackColor = [System.Drawing.Color]::FromArgb(50, 50, 50)
$btnCopy.Size   = New-Object System.Drawing.Size(90, 26)
$btnCopy.Location = New-Object System.Drawing.Point(8, 5)

$tb = New-Object System.Windows.Forms.TextBox
$tb.Multiline   = $true
$tb.ScrollBars  = 'Vertical'
$tb.Dock        = 'Fill'
$tb.ReadOnly    = $true
$tb.WordWrap    = $false
$tb.Font        = New-Object System.Drawing.Font('Consolas', 9)
$tb.BackColor   = [System.Drawing.Color]::FromArgb(18, 18, 18)
$tb.ForeColor   = [System.Drawing.Color]::FromArgb(200, 200, 200)
$tb.Text        = 'LOGCONTENT'

$btnCopy.Add_Click({ [System.Windows.Forms.Clipboard]::SetText($tb.Text) })

# scroll to bottom
$tb.Add_VisibleChanged({
    $tb.SelectionStart  = $tb.Text.Length
    $tb.SelectionLength = 0
    $tb.ScrollToCaret()
})

$panel.Controls.Add($btnCopy)
$form.Controls.Add($tb)
$form.Controls.Add($panel)
$form.ShowDialog()
`
	script = strings.ReplaceAll(script, "LOGCONTENT", content)

	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command", script)
	if err := cmd.Start(); err != nil {
		log.Printf("logwin: failed to open: %v", err)
	}
}
