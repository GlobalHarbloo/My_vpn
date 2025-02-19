package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"time"

	"gioui.org/app"
	"gioui.org/font/gofont"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/google/uuid"
)

var (
	usernameInput widget.Editor
	emailInput    widget.Editor
	passwordInput widget.Editor
	registerBtn   widget.Clickable
	loginBtn      widget.Clickable
	connectBtn    widget.Clickable
	statusLabel   widget.Label
)

func main() {
	go func() {
		w := app.NewWindow(
			app.Title("VPN Client"),
			app.Size(unit.Dp(400), unit.Dp(600)),
		)
		if err := loop(w); err != nil {
			fmt.Println("Error:", err)
		}
		os.Exit(0)
	}()
	app.Main()
}

func loop(w *app.Window) error {
	th := material.NewTheme(gofont.Collection())
	var ops layout.Ops

	for {
		e := <-w.Events()
		switch e := e.(type) {
		case system.DestroyEvent:
			return e.Err
		case system.FrameEvent:
			gtx := layout.NewContext(&ops, e)
			layout.Flex{
				Axis:    layout.Vertical,
				Spacing: layout.SpaceEvenly,
			}.Layout(gtx,
				layout.Rigid(material.H1(th, "VPN Client").Layout),
				layout.Rigid(material.Editor(th, &usernameInput, "Username").Layout),
				layout.Rigid(material.Editor(th, &emailInput, "Email").Layout),
				layout.Rigid(material.Editor(th, &passwordInput, "Password").Layout),
				layout.Rigid(material.Button(th, &registerBtn, "Register").Layout),
				layout.Rigid(material.Button(th, &loginBtn, "Login").Layout),
				layout.Rigid(material.Button(th, &connectBtn, "Connect to VPN").Layout),
				layout.Rigid(material.Label(th, unit.Dp(14), statusLabel.Text).Layout),
			)
			e.Frame(gtx.Ops)

			if registerBtn.Clicked() {
				go register()
			}
			if loginBtn.Clicked() {
				go login()
			}
			if connectBtn.Clicked() {
				go connectVPN()
			}
		}
	}
}

func register() {
	uuid := uuid.New().String()
	resp, err := http.PostForm("http://your_server_ip/register", map[string][]string{
		"username": {usernameInput.Text()},
		"email":    {emailInput.Text()},
		"password": {passwordInput.Text()},
		"uuid":     {uuid},
	})
	if err != nil {
		statusLabel.Text = "Registration failed!"
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		statusLabel.Text = "Registration successful! UUID: " + uuid
	} else {
		statusLabel.Text = "Registration failed!"
	}
}

func login() {
	resp, err := http.PostForm("http://your_server_ip/login", map[string][]string{
		"identifier": {usernameInput.Text()},
		"password":   {passwordInput.Text()},
	})
	if err != nil {
		statusLabel.Text = "Login failed!"
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		statusLabel.Text = "Login successful!"
	} else {
		statusLabel.Text = "Login failed!"
	}
}

func connectVPN() {
	cmd := exec.Command("v2ray", "-config=config.json")
	err := cmd.Start()
	if err != nil {
		statusLabel.Text = "Failed to start VPN client!"
		return
	}
	statusLabel.Text = "VPN client started successfully!"
	time.Sleep(2 * time.Second)
	cmd.Wait()
}
