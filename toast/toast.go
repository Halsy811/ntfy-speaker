package toast

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"text/template"

	"github.com/google/uuid"
)

var toastTemplate *template.Template

var (
	ErrorInvalidAudio    error = errors.New("toast: invalid audio")
	ErrorInvalidDuration       = errors.New("toast: invalid duration")
)

type toastAudio string

const (
	Default        toastAudio = "ms-winsoundevent:Notification.Default"
	IM             toastAudio = "ms-winsoundevent:Notification.IM"
	Mail           toastAudio = "ms-winsoundevent:Notification.Mail"
	SMS            toastAudio = "ms-winsoundevent:Notification.SMS"
	LoopingAlarm   toastAudio = "ms-winsoundevent:Notification.Looping.Alarm"
	LoopingAlarm2  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm2"
	LoopingAlarm3  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm3"
	LoopingAlarm4  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm4"
	LoopingAlarm5  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm5"
	LoopingAlarm6  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm6"
	LoopingAlarm7  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm7"
	LoopingAlarm8  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm8"
	LoopingAlarm9  toastAudio = "ms-winsoundevent:Notification.Looping.Alarm9"
	LoopingAlarm10 toastAudio = "ms-winsoundevent:Notification.Looping.Alarm10"
	LoopingCall    toastAudio = "ms-winsoundevent:Notification.Looping.Call"
	LoopingCall2   toastAudio = "ms-winsoundevent:Notification.Looping.Call2"
	LoopingCall3   toastAudio = "ms-winsoundevent:Notification.Looping.Call3"
	LoopingCall4   toastAudio = "ms-winsoundevent:Notification.Looping.Call4"
	LoopingCall5   toastAudio = "ms-winsoundevent:Notification.Looping.Call5"
	LoopingCall6   toastAudio = "ms-winsoundevent:Notification.Looping.Call6"
	LoopingCall7   toastAudio = "ms-winsoundevent:Notification.Looping.Call7"
	LoopingCall8   toastAudio = "ms-winsoundevent:Notification.Looping.Call8"
	LoopingCall9   toastAudio = "ms-winsoundevent:Notification.Looping.Call9"
	LoopingCall10  toastAudio = "ms-winsoundevent:Notification.Looping.Call10"
	Silent         toastAudio = "silent"
)

type toastDuration string

const (
	Short toastDuration = "short"
	Long  toastDuration = "long"
)

type toastScenario string

const (
	Reminder     toastScenario = "reminder"
	Alarm        toastScenario = "alarm"
	IncomingCall toastScenario = "incomingCall"
	Urgent       toastScenario = "urgent"
)

func init() {
	toastTemplate = template.New("toast")
	var err error
	toastTemplate, err = toastTemplate.Parse(`
[Windows.UI.Notifications.ToastNotificationManager, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.UI.Notifications.ToastNotification, Windows.UI.Notifications, ContentType = WindowsRuntime] | Out-Null
[Windows.Data.Xml.Dom.XmlDocument, Windows.Data.Xml.Dom.XmlDocument, ContentType = WindowsRuntime] | Out-Null

$APP_ID = '{{if .AppID}}{{.AppID}}{{else}}Windows App{{end}}'

$template = @"
<toast activationType="{{.ActivationType}}" launch="{{.ActivationArguments}}" duration="{{.Duration}}" {{if .Scenario}}scenario="{{.Scenario}}"{{end}}>
    <visual>
        <binding template="ToastGeneric">
            {{if .Icon}}
            <image placement="appLogoOverride" src="{{.Icon}}" />
            {{end}}
            {{if .Title}}
            <text><![CDATA[{{.Title}}]]></text>
            {{end}}
            {{if .Message}}
            <text><![CDATA[{{.Message}}]]></text>
            {{end}}
        </binding>
    </visual>
    {{if ne .Audio "silent"}}
    <audio src="{{.Audio}}" loop="{{.Loop}}" />
    {{else}}
    <audio silent="true" />
    {{end}}
    <actions>
        {{if eq .Scenario "reminder"}}
        <action activationType="system" arguments="dismiss" content="Закрыть" id="dismiss" />
        {{end}}
        {{if .Actions}}
        {{range .Actions}}
        <action activationType="{{.Type}}" content="{{.Label}}" arguments="{{.Arguments}}" />
        {{end}}
        {{end}}
    </actions>
</toast>
"@

$xml = New-Object Windows.Data.Xml.Dom.XmlDocument
$xml.LoadXml($template)
$toast = New-Object Windows.UI.Notifications.ToastNotification $xml
[Windows.UI.Notifications.ToastNotificationManager]::CreateToastNotifier($APP_ID).Show($toast)
    `)
	if err != nil {
		panic(err)
	}
}

type Notification struct {
	AppID               string
	Title               string
	Message             string
	Icon                string
	ActivationType      string
	ActivationArguments string
	Actions             []Action
	Audio               toastAudio
	Loop                bool
	Duration            toastDuration
	Scenario            toastScenario
}

type Action struct {
	Type      string
	Label     string
	Arguments string
}

func (n *Notification) applyDefaults() {
	if n.ActivationType == "" {
		n.ActivationType = "protocol"
	}
	if n.Duration == "" {
		n.Duration = Short
	}
	if n.Audio == "" {
		n.Audio = Default
	}
}

func (n *Notification) buildXML() (string, error) {
	var out bytes.Buffer
	err := toastTemplate.Execute(&out, n)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (n *Notification) Push() error {
	n.applyDefaults()
	xml, err := n.buildXML()
	if err != nil {
		return err
	}
	return invokeTemporaryScript(xml)
}

func Audio(name string) (toastAudio, error) {
	switch strings.ToLower(name) {
	case "default":
		return Default, nil
	case "im":
		return IM, nil
	case "mail":
		return Mail, nil
	case "sms":
		return SMS, nil
	case "loopingalarm":
		return LoopingAlarm, nil
	case "loopingalarm2":
		return LoopingAlarm2, nil
	case "loopingalarm3":
		return LoopingAlarm3, nil
	case "loopingalarm4":
		return LoopingAlarm4, nil
	case "loopingalarm5":
		return LoopingAlarm5, nil
	case "loopingalarm6":
		return LoopingAlarm6, nil
	case "loopingalarm7":
		return LoopingAlarm7, nil
	case "loopingalarm8":
		return LoopingAlarm8, nil
	case "loopingalarm9":
		return LoopingAlarm9, nil
	case "loopingalarm10":
		return LoopingAlarm10, nil
	case "loopingcall":
		return LoopingCall, nil
	case "loopingcall2":
		return LoopingCall2, nil
	case "loopingcall3":
		return LoopingCall3, nil
	case "loopingcall4":
		return LoopingCall4, nil
	case "loopingcall5":
		return LoopingCall5, nil
	case "loopingcall6":
		return LoopingCall6, nil
	case "loopingcall7":
		return LoopingCall7, nil
	case "loopingcall8":
		return LoopingCall8, nil
	case "loopingcall9":
		return LoopingCall9, nil
	case "loopingcall10":
		return LoopingCall10, nil
	case "silent":
		return Silent, nil
	default:
		return Default, ErrorInvalidAudio
	}
}

func Duration(name string) (toastDuration, error) {
	switch strings.ToLower(name) {
	case "short":
		return Short, nil
	case "long":
		return Long, nil
	default:
		return Short, ErrorInvalidDuration
	}
}

func invokeTemporaryScript(content string) error {
	id := uuid.New()
	file := filepath.Join(os.TempDir(), id.String()+".ps1")
	defer os.Remove(file)
	bomUtf8 := []byte{0xEF, 0xBB, 0xBF}
	out := append(bomUtf8, []byte(content)...)
	err := os.WriteFile(file, out, 0600)
	if err != nil {
		return err
	}
	cmd := exec.Command("PowerShell", "-ExecutionPolicy", "Bypass", "-File", file)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	if err = cmd.Run(); err != nil {
		return err
	}
	return nil
}
