package main

import (
	"basher"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"os"
	"ssher"
	"strings"
	"time"
)

var command1 = []string{"crm", "status"}
var Nodes, Unregistered []string

//-----checks if its sles11sp4 then
func DistroChecker(ip string) string {
	commandx := "cat /etc/os-release | grep SLE"
	checker := ssher.SSH("root", ip, commandx, "string")
	if strings.Contains(checker, "11") && strings.Contains(checker, "SP4") {
		fmt.Println("The distro is 11, not sure haproxy is in there")
	}
	return "11sp4"
}

func Regchecker(command2 string, ip string) int {
	var checker []string
	//-----------checking if Nodes are registered on sbd:
	checker = strings.Split(ssher.SSH("root", ip, command2, "string"), "\n")
	var registered int
	for i := 0; i < len(Nodes); i++ {
		Unregistered = append(Unregistered, Nodes[i])
	}
	//making an identical copy of Nodes into Unregistered slice and checking if they appear in sbd list output... if yes - delete the node name from Unregistered
	//until the slice is empty

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
	ip := "10.160.67.101"
	var sbddevice string
	var installed int

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

	time.Sleep(2 * time.Second)
	tmp := strings.Split(sbddevice, "=")
	command2 := "sbd -d " + tmp[len(tmp)-1] + " list"
	registered := Regchecker(command2, ip)

	//----------if not all three nodes are registered to the SBD ---> the program does it "manually" by allocating slot for each node
	if registered < 3 {
		fmt.Printf("\nd. Not all nodes registered, performing node registering \"manually\"...\nhere is the list of unregistered nodes: %v\n", Unregistered)
		for i := 0; i < len(Unregistered); i++ {
			command3 := "sbd -d " + sbddevice + " allocate " + Unregistered[i]
			ssher.SSH("root", ip, command3, "print")
		}
		registered = Regchecker(command2, ip)
	} else {
		fmt.Println("\ne. All nodes well-registered on running SBD and ready!\n")
	}

	//------checking if haproxy installed, if not - it will install it
	command4 := "zypper se -s haproxy"
	if DistroChecker(ip) != "11sp4" {
		checker = strings.Split(ssher.SSH("root", ip, command4, "string"), "\n")
		for i := 0; i < len(checker); i++ {
			if strings.Contains(checker[i], "i+") {
				installed++
				fmt.Println("\nf. haproxy is installed!\n")
			}
		}
	}

	//------installing haproxy if its not on the machine
	if installed < 1 {
		checker = strings.Split(ssher.SSH("root", ip, "zypper -n install haproxy", "string"), "\n")
	}
	//fmt.Println(checker[len(checker)-2])
	if strings.Contains(checker[len(checker)-2], "1/1") && strings.Contains(checker[len(checker)-2], "Installing") && strings.Contains(checker[len(checker)-2], "done") {
		fmt.Println("\nf. Succes! haproxy successfully installed!")
	}

	tpl, err := template.ParseFiles("haproxytemplate")
	if err != nil {
		log.Fatalf("Search didn't work...%s", err)
	}

	//fmt.Printf("here is the Nodes slice: %v", Nodes)
	haproxycfg := map[string]interface{}{
		"Node3":          Nodes[len(Nodes)-3],
		"Node1":          Nodes[len(Nodes)-2],
		"Node2":          Nodes[len(Nodes)-1],
		"Ipandport":      ip + ":80",
		"Ipandportnode1": "10.160.66.241:80",
		"Ipandportnode2": "10.160.67.101:80",
	}

	//creating a new file named a
	command5 := []string{"echo", "\"\"", ">", "a"}
	basher.Bash(command5, "s")

	//writer, err := os.OpenFile("a", syscall.O_RDONLY|syscall.O_WRONLY, 0644)
	//if err != nil {
	//	log.Fatalf("couldn't open the file...%s", err)
	//}

	var f *os.File
	f, err = os.Create("a")
	if err != nil {
		log.Fatalf("couldn't create the file...%s", err)
	}

	//---executing the template, putting result into an *os.File, e.g. into an opened new File
	err = tpl.Execute(f, haproxycfg)
	if err != nil {
		log.Fatalf("couldn't execute the report...%s", err)
	}

	//-----reading the freshly created interpreted config file in order to convert
	//----it into string and include into the ECHO ssh command on one of the nodes
	tmp2, err := ioutil.ReadFile("a")
	str := string(tmp2)
	//fmt.Println(str)

	//--- $str is the whole config file
	command6 := "echo " + "'" + str + "' >/etc/haproxy/haproxy.cfg"
	//fmt.Println(command6)
	ssher.SSH("root", ip, command6, "string")
	if err == nil {
		command9 := "cat /etc/haproxy/haproxy.cfg"
		resp := strings.Split(ssher.SSH("root", ip, command9, "string"), " ")
		var count int
		for i := 0; i < len(resp); i++ {
			if strings.Contains(resp[i], "haproxy.cfg") || strings.Contains(resp[i], Nodes[len(Nodes)-1]) || strings.Contains(resp[i], Nodes[len(Nodes)-2]) || strings.Contains(resp[i], Nodes[len(Nodes)-3]) {
				count += 1
			}
		}
		if count >= 3 {
			fmt.Println("g. haproxy config file successfully uploaded")
		}
	}
	//------UPLOADED the haproxy config! DONE!

	time.Sleep(2 * time.Second)

	//-----Setting up a csync2 config profile red from a predefinde template file (csync2template)
	tpl, err = template.ParseFiles("csync2template")
	if err != nil {
		log.Fatalf("Search didn't work...%s", err)
	}

	f, err = os.Create("b")
	if err != nil {
		log.Fatalf("couldn't create the file...%s", err)
	}

	err = tpl.Execute(f, haproxycfg)
	if err != nil {
		log.Fatalf("couldn't execute the report...%s", err)
	}

	tmp2, err = ioutil.ReadFile("b")
	str = string(tmp2)

	command7 := "echo " + "'" + str + "' >/etc/csync2/csync2.cfg"
	ssher.SSH("root", ip, command7, "string")

	if err == nil {
		var count int
		command8 := "cat /etc/csync2/csync2.cfg"
		resp := strings.Split(ssher.SSH("root", ip, command8, "string"), "\n")
		for i := 0; i < len(resp); i++ {
			if strings.Contains(resp[i], "haproxy.cfg") || strings.Contains(resp[i], Nodes[len(Nodes)-1]) || strings.Contains(resp[i], Nodes[len(Nodes)-2]) || strings.Contains(resp[i], Nodes[len(Nodes)-3]) {
				count += 1
			}
		}

		if count >= 4 {
			fmt.Println("h. csync2 config file successfully uploaded")
		}
	}
	//------UPLOADED the csync2 config! DONE!

	//----syncing the csync2
	command10 := "csync2 -xv " + ">" + "&" + "log"
	ssher.SSH("root", ip, command10, "print")

	time.Sleep(10 * time.Second)
	command11 := "cat log"
	ssher.SSH("root", ip, command11, "print")

	//fmt.Println(resp)
}
