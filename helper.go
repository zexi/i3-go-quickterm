package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"go.i3wm.org/i3/v4"

	"yunion.io/x/jsonutils"
)

func GetCurrentWorkspace() (*i3.Workspace, error) {
	wss, err := i3.GetWorkspaces()
	if err != nil {
		return nil, err
	}
	for _, ws := range wss {
		if ws.Focused {
			return &ws, nil
		}
	}
	return nil, fmt.Errorf("not found focused workspace")
}

func MoveBack(selector string) error {
	cmd := fmt.Sprintf("%s floating enable, move scratchpad", selector)
	_, err := i3.RunCommand(cmd)
	return err
}

func GetCurrentOutput() (*i3.Output, error) {
	outs, err := i3.GetOutputs()
	if err != nil {
		return nil, fmt.Errorf("get outputs: %v", err)
	}
	for _, out := range outs {
		if out.Active {
			return &out, nil
		}
	}
	return nil, fmt.Errorf("not found current output")
}

func PopIt(markName string, position string, ratio float32) error {
	ws, err := GetCurrentWorkspace()
	if err != nil {
		return fmt.Errorf("get current workspace: %v", err)
	}
	wx := ws.Rect.X
	wy := ws.Rect.Y
	wwidth := ws.Rect.Width
	wheight := ws.Rect.Height
	height := int64(float32(wheight) * ratio)
	var posx int64 = wx
	var posy int64 = wy
	if position == "bottom" {
		var margin int64 = 6
		posy = wy + wheight - height - margin
	}
	cmds := []string{
		fmt.Sprintf("[con_mark=%s]", markName),
		"move scratchpad,",
		"scratchpad show,",
		fmt.Sprintf("resize set %d px %d px,", wwidth, height),
		fmt.Sprintf("move absolute position %dpx %dpx", posx, posy),
	}
	_, err = i3.RunCommand(strings.Join(cmds, ""))
	return err
}

func ExitErr(err error) {
	if err != nil {
		log.Printf("Error: %v", err)
		os.Exit(1)
	}
	os.Exit(0)
}

func IsWorkspaceContainsNode(wsNode *i3.Node, node *i3.Node) (bool, error) {
	exists := false
	wsNode.FindChild(func(child *i3.Node) bool {
		if child.ID == node.ID {
			exists = true
			return true
		}
		return false
	})
	return exists, nil
}

func GetWorkspaceNodes() ([]*i3.Node, error) {
	tree, err := i3.GetTree()
	if err != nil {
		return nil, err
	}
	workspaces := make([]*i3.Node, 0)
	tree.Root.FindChild(func(node *i3.Node) bool {
		if node.Type == i3.WorkspaceNode {
			workspaces = append(workspaces, node)
		}
		return false
	})
	return workspaces, nil
}

func FindNodeWorkspace(node *i3.Node) (*i3.Node, error) {
	wss, err := GetWorkspaceNodes()
	if err != nil {
		return nil, fmt.Errorf("get workspaces: %v", err)
	}
	log.Printf(fmt.Sprintf("Get workspace nodes: %s", jsonutils.Marshal(wss).PrettyString()))
	for _, ws := range wss {
		if found, err := IsWorkspaceContainsNode(ws, node); found {
			return ws, nil
		} else {
			log.Printf("Not found %v", err)
		}
	}
	return nil, fmt.Errorf("not found workspace contain node")
}
