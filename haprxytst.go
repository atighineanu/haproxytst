package main

import (
	"fmt"
	"ssher"
	"strings"
)

var command1 = []string{"crm", "status"}

func main() {
	ip := "10.160.66.220"
	var sbddevice string
	var nodes []string
	//-----running crm status to check if nodes online
	checker := strings.Split(ssher.SSH("root", ip, "crm status", "print"), "\n")
	for i := 0; i < len(checker); i++ {
		if strings.Contains(checker[i], "Online") {
			line := strings.Split(checker[i], " ")
			for j := 0; j < len(line); j++ {
				if line[j] == "[" {
					//-------------checking if there are three nodes between the brackets [ 1 2 3 ]
					//-------------remembering the node titles "1" "2" "3" for further testing into array "nodes"
					for k := 1; k <= 3; k++ {
						if line[j+k] != "]" {
							nodes = append(nodes, line[j+k])
							if len(nodes) > 2 {
								fmt.Printf("a. Three-node cluster set up and running... \n %v\n", nodes)
							}
						}
					}
				}
			}
		}

		//---------checking if stonith-sbd runs properly on ANY of the nodes
		if strings.Contains(checker[i], "stonith-sbd") && strings.Contains(checker[i], "Started") {
			for k := 0; k < len(nodes); k++ {
				if strings.Contains(checker[i], nodes[k]) {
					fmt.Printf("\nb. Stonith-SBD runs properly on one of the nodes: \n%v (node name)\n", nodes[k])
				}
			}
		}

	}

	//-----------checking the SBD device, and checking IF all nodes on SBD are registered.............
	checker = strings.Split(ssher.SSH("root", ip, "cat /etc/sysconfig/sbd", "string"), "\n")
	for i := 0; i < len(checker); i++ {
		if strings.Contains(checker[i], "SBD_DEVICE") {
			line := strings.Split(checker[i], "\"")
			for j := 0; j < len(line); j++ {
				if strings.Contains(line[j], "/") {
					fmt.Printf("\nc. SBD runs on this device: %v\n", line[j])
					sbddevice = line[j]
				}
			}
		}

	}
	//-----------checking if Nodes are registered on sbd:
	command2 := "sbd -d " + sbddevice + " list"
	checker = strings.Split(ssher.SSH("root", ip, command2, "string"), "\n")
	for 

}
