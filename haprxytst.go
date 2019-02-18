package main

import (
	"fmt"
	"ssher"
	"strings"
)

var command1 = []string{"crm", "status"}
var Nodes, Unregistered []string

func Regchecker(command2 string, ip string) int {
	//-----------checking if Nodes are registered on sbd:

	checker := strings.Split(ssher.SSH("root", ip, command2, "string"), "\n")
	fmt.Println(checker)
	var registered int
	Unregistered := Nodes
	for i := 0; i < len(checker); i++ {
		for j := 0; j < len(Nodes); j++ {
			if strings.Contains(checker[i], Nodes[j]) {
				registered += 1
				for k := 0; k < len(Unregistered); k++ {
					if Nodes[j] == Unregistered[k] {
						copy(Unregistered[k:], Unregistered[k+1:]) // Shift Unregistered[k+1:] left one index.
						Unregistered[len(Unregistered)-1] = ""     // Erase last element (write zero value).
						Unregistered = Unregistered[:len(Unregistered)-1]
					}
				}
			}
		}
	}
	return registered
}

//func haproxytester(command4 string, ip string) {
//	checker := strings.Split(ssher.SSH("root", ip, command2, "string"), "\n")
//}

func main() {
	ip := "10.160.66.220"
	var sbddevice string

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
							Nodes = append(Nodes, line[j+k])
							if len(Nodes) > 2 {
								fmt.Printf("a. Three-node cluster set up and running... \n %v\n", Nodes)
							}
						}
					}
				}
			}
		}

		//---------checking if stonith-sbd runs properly on ANY of the nodes
		if strings.Contains(checker[i], "stonith-sbd") && strings.Contains(checker[i], "Started") {
			for k := 0; k < len(Nodes); k++ {
				if strings.Contains(checker[i], Nodes[k]) {
					fmt.Printf("\nb. Stonith-SBD runs properly on one of the nodes: \n%v (node name)\n", Nodes[k])
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

	command2 := "sbd -d " + sbddevice + " list"
	registered := Regchecker(command2, ip)

	//----------if not all three nodes are registered to the SBD ---> the program does it "manually" by allocating slot for each node
	if registered <= 3 {
		fmt.Printf("\nd. Not all nodes registered, performing node registering \"manually\"...\nhere is the list of unregistered nodes: %v\n", Unregistered)
		for i := 0; i < len(Unregistered); i++ {
			command3 := "sbd -d " + sbddevice + " allocate " + Unregistered[i]
			ssher.SSH("root", ip, command3, "print")
		}
		registered = Regchecker(command2, ip)
	} else {
		fmt.Println("\ne. All nodes well-registered on running SBD and ready!\n")
	}

	//------checking if haproxy installed....
	command4 := "rpm -qi haproxy"
	ssher.SSH("root", ip, command4, "print")
	// TO DO... see why
}
