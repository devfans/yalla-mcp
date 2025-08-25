package main

import (
	"context"
	"crypto/md5"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/devfans/envconf/dotenv"
	"github.com/devfans/golang/log"
	"github.com/google/uuid"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// Global variables
var (
	DeviceID = genDeviceID()
	AppID = genAppID()
	AppSecret = genSecret()
)


const NOTES = `
NOTES:
- 走廊连接着客厅，厨房，玄关，主卧，次卧和卫生间
- 吊灯在主卧, 左灯，右灯分别在主卧床的两侧 
- Button "客厅打开" 会打开客厅所有灯光, 次卧打开/卫生间打开/厨房打开/玄关打开/主卧打开 同理，以及对应的关闭按钮
- 桌面是客厅的一部分，只有灯带，氛围灯也在客厅
- 客厅灯带包含 桌面灯带和电视灯带
- 餐桌灯在桌面旁边，但餐桌在走廊，吃饭时需要走廊灯和厨房灯但不需要餐桌灯
`

const (
	Version                         = "0.0.3"
	RequestSignatureHeaderAccessKey = "X-Access-Key"
	RequestSignatureHeaderSignature = "X-Signature"
	RequestSignatureHeaderTimestamp = "X-Timestamp"
	RequestSignatureHeaderNonce     = "X-Nonce"
	DefaultAPITimeout               = 10 * time.Second
	DefaultAPPTimeout               = 15 * time.Second
)

var (
	API_BASE_URL = "https://ai-echo.aqara.cn/echo/mcp"
	API_KEY = dotenv.String("API_KEY")
	API_TOKEN = dotenv.String("API_TOKEN")
)

func genSecret() string {
	url := API_BASE_URL + "/secret"
	result, err := httpGet[map[string]string](url, map[string]string{"key": AppID})
	if err != nil {
		log.Error("Failed to generate secret", "err", err)
		return ""
	}
	if result == nil {
		log.Warn("No secret returned from server")
		return ""
	}
	if v, ok := (*result)["secret_key"]; ok {
		return v
	}
	log.Warn("Secret key not found in response")
	return ""
}

// genDeviceID generates a unique device identifier.
func genDeviceID() string {
	var macAddr string
	interfaces, err := net.Interfaces()
	if err == nil {
		for _, i := range interfaces {
			if i.Flags&net.FlagUp != 0 && !strings.HasPrefix(i.Name, "lo") && len(i.HardwareAddr) > 0 {
				macAddr = i.HardwareAddr.String()
				break
			}
		}
	}

	prefix := "mcp0."
	if macAddr == "" {
		macAddr = uuid.NewString()
		prefix = "mcp1."
	}

	hostname, _ := os.Hostname()
	osInfo := runtime.GOOS + "-" + runtime.GOARCH

	baseInfo := strings.Join([]string{macAddr, hostname, osInfo}, "-")
	hash := sha1.New()
	hash.Write([]byte(baseInfo))

	return prefix + hex.EncodeToString(hash.Sum(nil))
}

// genAppID generates an application identifier.
func genAppID() string {
	prefix := "mcp-"
	return prefix + md5Hash(prefix+DeviceID)
}

func md5Hash(str string) string {
	hasher := md5.New()
	hasher.Write([]byte(str))
	return hex.EncodeToString(hasher.Sum(nil))
}


// Using the generic AddTool automatically populates the the input and output
// schema of the tool.
//
// The schema considers 'json' and 'jsonschema' struct tags to get argument
// names and descriptions.
var list_home = &mcp.Tool{
	Name:        "list_homes",
	Description: `Get all homes under the user (useful when the user wants to query/switch homes).
Returns:
Comma-separated list of home names; returns an empty string or specific message if no data.
`,
}

func HandleListHome(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	log.Info("GetHomesHandler request", "args", args)
	homes, message := GetHomes()
	if message != "" {
		log.Error("GetHomes failed", "message", message)
		return simpleResult(message), nil, nil
	}
	log.Info("Home list retrieved", "homes", homes)
	if len(homes) == 0 {
		return simpleResult("No homes found."), nil, nil
	}
	return simpleResult(homes...), nil, nil
}

type args struct {
	Name string `json:"name" jsonschema:"the person to greet"`
}

var switch_home = &mcp.Tool{
	Name:        "switch_home",
	Description: `Switch the user's current home.
Returns:
Switch result message.
`,
}

func HandleSwitchHome(ctx context.Context, req *mcp.CallToolRequest, args args) (*mcp.CallToolResult, any, error) {
	log.Info("SwitchHomeHandler request", "args", args)
	log.Info("Switching home", "homeName", args.Name)
	success, message := SwitchHome(args.Name)
	if !success {
		log.Error("Home switch failed", "message", message)
		// Ensure a message is always returned on failure.
		if message == "" {
			message = "Home switch failed due to an unknown error."
		}
		return simpleResult(message), nil, nil
	}
	log.Info("Switched to home", "homeName", args.Name)
	return simpleResult(fmt.Sprintf("Successfully switched to home \"%s\"", args.Name)), nil, nil
}

var list_scenes = &mcp.Tool{
	Name:        "list_device_control_buttons",
	Description: `Get all device control buttons under the user's home.
Returns:
  Control buttons information in Markdown format` + NOTES,
}

// GetScenesHandler handles querying available scenes.
func HandleListScenesHandler(ctx context.Context, req *mcp.CallToolRequest, args struct{}) (*mcp.CallToolResult, any, error) {
	log.Info("GetScenesHandler request", "args", req.Params.Arguments)
	result := GetScenes([]string{})
	result = strings.ReplaceAll(result, "scene", "device button")
	log.Info("GetScenes result", "result", result)
	return simpleResult(result), nil, nil
}

var run_scenes = &mcp.Tool{
	Name:        "push_device_control_button",
	Description: `Push device control buttons under the user's home, or control buttons in a specified room.
Returns:
  Device control button push result message.`,
}
type argScenes struct {
	Button int `json:"button" jsonschema:"the control button to push, exactly one button should be provided"`
}
// GetScenesHandler handles querying available scenes.
func HandleRunScenesHandler(ctx context.Context, req *mcp.CallToolRequest, args argScenes) (*mcp.CallToolResult, any, error) {
	log.Info("HandleRunScenesHandler request", "args", args)
	log.Info("Running scene", "button", args.Button)
	result := RunScenes([]int{args.Button})
	log.Info("RunScene result", "result", result)
	return simpleResult(result), nil, nil
}

func registerTools(server *mcp.Server) {
	// mcp.AddTool(server, list_home, HandleListHome);
	// mcp.AddTool(server, switch_home, HandleSwitchHome)
	a, b := SwitchHome("我的家")
	log.Info("Switching home", a, b)
	mcp.AddTool(server, list_scenes, HandleListScenesHandler)
	mcp.AddTool(server, run_scenes, HandleRunScenesHandler)
}