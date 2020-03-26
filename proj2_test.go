package proj2

// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder, and we will be Very Upset.

import (
	"testing"
	"reflect"
	//"fmt"
	//"time"
	"github.com/cs161-staff/userlib"
	_ "encoding/json"
	_ "encoding/hex"
	"github.com/google/uuid"
	"strings"
	_ "errors"
	_ "strconv"
)

func clear() {
	// Wipes the storage so one test does not affect another
	userlib.DatastoreClear()
	userlib.KeystoreClear()
}

func TestInit(t *testing.T) {
	clear()

	// You can set this to false!
	userlib.SetDebugStatus(true)

	// 1. Init test
	u, err := InitUser("a", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}

	// 2. Duplicate name test
	_, err = InitUser("a", "foobar")
	if err.Error() != strings.ToTitle("User already exists") {
		t.Error("Failed to ban duplicate user init", err)
		return
	}

	// 3. Get user test
	u2, err := GetUser("a", "fubar")
	if err != nil {
		t.Error("Failed to support 2nd-instantiation of user", err)
		return
	}

	u3, err := GetUser("a", "fubar")
	if err != nil {
		t.Error("Failed to support 3rd-instantiation of user", err)
		return
	}

	// 4. Non-existence user test
	_, err = GetUser("b", "foo")
	if err == nil {
		t.Error("Failed non-existence user query test", err)
		return
	}

	// 5. log in verification test
	_, err = GetUser("a", "foobar")
	if err.Error() != strings.ToTitle("Incorrect username or password") {
		t.Error("Failed user authentication test", err)
		return
	}

	_ = u
	_ = u2
	_ = u3
	return

}

func TestTamper(t *testing.T) {
	clear()
	u, err := InitUser("a", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	dmap := userlib.DatastoreGetMap()
	for k, _ := range dmap {
		userlib.DatastoreSet(k, []byte("dwjedgwvcuaghdisacyukgwefiukcghweviasfcvwekugasfciwuehawgsdif"))
	}

	_, err = GetUser("a", "fubar")
	if err == nil {
		t.Error("Ftampered not detected", err)
		return
	}

	clear()
	u, err = InitUser("a", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	dmap1 := []uuid.UUID{}
	for k,_ := range userlib.DatastoreGetMap() {
  	dmap1 = append(dmap1, k)
	}

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	dmap2 := userlib.DatastoreGetMap()


	for k2, _ := range dmap2 {
			contain := false
			for _, k1 := range dmap1 {
				if k1 == k2 {
					contain = true
				}
		}
		if contain == false {
			userlib.DatastoreSet(k2, []byte("dwjedgwvcuaghdisacyukgwefiukcghweviasfcvwekugasfciwuehawgsdif"))
		}
	}

	v, err = u.LoadFile("file1")
	if err == nil {
		t.Error("Ftampered not detected", string(v))
		return
	}

	clear()
	u, err = InitUser("a", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	v = []byte("This is a test")
	u.StoreFile("file1", v)
	dmap1 = []uuid.UUID{}
	for k,_ := range userlib.DatastoreGetMap() {
  	dmap1 = append(dmap1, k)
	}
	u.AppendFile("file1", []byte("kid you not"))

	dmap2 = userlib.DatastoreGetMap()
	for k2, _ := range dmap2 {
			//fmt.Println(dmap1[k2])
			contain := false
			for _, k1 := range dmap1 {
				if k1 == k2 {
					contain = true
				}
		}
		if contain == false {
			userlib.DatastoreSet(k2, []byte("dwjedgwvcuaghdisacyukgwefiukcghweviasfcvwekugasfciwuehawgsdif"))
		}
	}

	clear()
	u1, err := InitUser("a", "fubar")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	u2, err := InitUser("b", "foob")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to initialize user", err)
		return
	}
	v = []byte("This is a test")
	u.StoreFile("file1", v)
	dmap3 := []uuid.UUID{}
	for k,_ := range userlib.DatastoreGetMap() {
  	dmap3 = append(dmap3, k)
	}
	magic_string, err := u1.ShareFile("file1", "b")
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to share file", err)
		return
	}
	err = u2.ReceiveFile("file2", "a", magic_string)
	if err != nil {
		// t.Error says the test fails
		t.Error("Failed to receive the file", err)
		return
	}
	dmap4 := userlib.DatastoreGetMap()
	for k4, _ := range dmap4 {
			//fmt.Println(dmap1[k2])
			contain := false
			for _, k1 := range dmap1 {
				if k1 == k4 {
					contain = true
				}
		}
		if contain == false {
			userlib.DatastoreSet(k4, []byte("dwjedgwvcuaghdisacyukgwefiukcghweviasfcvwekugasfciwuehawgsdif"))
		}
	}
	_, err = u2.LoadFile("file2")
	if err == nil {
		t.Error("Tampered not detected", err)
		return
	}

}


func TestStorage(t *testing.T) {
	clear()

	//ti := time.Now()
	// 1. Simple upload and download test
	u, err := InitUser("a", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	////fmt.Println("instantiation complete", time.Since(ti))
	//ti =time.Now()

	v := []byte("This is a test")
	u.StoreFile("file1", v)
	//fmt.Println("short store complete", time.Since(ti))
	//ti =time.Now()

	vu, err2 := u.LoadFile("file1")
	//fmt.Println("short load complete", time.Since(ti))
	//ti =time.Now()
	if err2 != nil {
		t.Error("Failed to upload and download", err2)
		return
	}
	if !reflect.DeepEqual(v, vu) {
		t.Error("Downloaded file is not the same", v, vu)
		return
	}

	// 2. Non-existence file test
	_, err = u.LoadFile("file2")
	if err == nil {
		t.Error("Failed non-existence file query test", err)
		return
	}

	// 3. multi-instantiation upload and download test
	u2, err := GetUser("a", "fubar")
	if err != nil {
		t.Error("Failed to support 2nd-instantiation of user", err)
		return
	}

	vu2, err := u2.LoadFile("file1")
	if err != nil {
		t.Error("Failed to load file for 2nd-instantiation", err)
		return
	}
	if !reflect.DeepEqual(v, vu2) {
		t.Error("Downloaded file is not the same for 2nd-instantiation", v, vu2)
		return
	}

	// 4. same-user overwrite test
	//ti =time.Now()
	v2 := []byte(", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 	 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 	 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 	 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 	 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 	 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 	 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 		 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 		 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 		 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 		 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 		 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 		 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 			 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 			 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 			 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 			 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 			 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 			 "scesidcgisavscb"+
 			 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 				 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 				 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 				 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 				 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 				 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 				 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 					 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 					 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 					 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 					 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 					 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 					 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 						 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 						 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 						 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 						 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 						 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 						 "scesidcgisavscb"+
						 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
							 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
							 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
							 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
							 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
							 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
							 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
								 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
								 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
								 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
								 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
								 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
								 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
									 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
									 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
									 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
									 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
									 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
									 "scesidcgisavscb"+
									 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
										 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
										 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
										 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
										 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
										 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
										 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
											 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
											 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
											 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
											 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
											 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
											 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
												 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
												 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
												 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
												 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
												 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
												 "scesidcgisavscb")
	u2.StoreFile("file1", v2)
	//fmt.Println("long store complete", time.Since(ti))
	//ti =time.Now()
	v4, err4 := u2.LoadFile("file1")
	//fmt.Println("long load complete", time.Since(ti))
	//ti =time.Now()
	if err4 != nil {
		t.Error("Failed to upload and download after overwrite", err4)
		return
	}
	if !reflect.DeepEqual(v2, v4) {
		t.Error("Downloaded file is not the same after overwrite", v2, v4)
		return
	}

	// 5. overwrite update test
	v5, err5 := u.LoadFile("file1")
	if err5 != nil {
		t.Error("Failed to load for 1st-instantiation after overwrite", err5)
		return
	}
	if !reflect.DeepEqual(v2, v5) {
		t.Error("Downloaded file is not the same for 1st-instantiation after overwrite", v2, v5)
		return
	}

	u2.StoreFile("file2", v)
	v6, err6 := u2.LoadFile("file2")
	if err6 != nil {
		t.Error("Failed to load new file after u2 stored", err6)
		return
	}
	if !reflect.DeepEqual(v, v6) {
		t.Error("Downloaded file is not the same for 1st-instantiation after overwrite", v, v6)
		return
	}

	// 7. nil file name test
 v61 := []byte("Nil filename test")
 u2.StoreFile("", v61)
 v62, err7 := u2.LoadFile("")
 if !reflect.DeepEqual(v61, v62) {
  t.Error("Downloaded file is not the same for nil named file", v61, v62)
  return
 }

 if err7 != nil {
  t.Error("Failed to store a file named as nil", err7)
  return
 }

 // 8. nil file content test

 v81 := []byte("")
 u2.StoreFile("test8", v81)
 v82, err8 := u2.LoadFile("test8")
 if !reflect.DeepEqual(v81, v82) {
  t.Error("Downloaded file is not the same for nil file content", v81, v82)
  return
 }

 if err8 != nil {
  t.Error("Failed to store a nil file content", err8)
  return
 }

 // destroy test
 clear()
 _, err9 := u2.LoadFile("test8")
	if err9 == nil {
  t.Error("You got it back after clear()", err9)
  return
 }

	return
}


func TestAppend(t *testing.T) {
	// 1. Simple upload and download test
	clear()

	u, err := InitUser("a", "fubar")
	if err != nil {
		t.Error("Failed to initialize user", err)
		return
	}
	v := []byte("This is a test")
	u.StoreFile("file1", v)


	// 2. Simple append test
	//ti := time.Now()
	v2 := []byte(", kid you not")
	u.AppendFile("file1", v2)
	//fmt.Println("short append complete", time.Since(ti))
	//ti =time.Now()
	vu2, err2 := u.LoadFile("file1")
	//fmt.Println("short load complete", time.Since(ti))
	//ti =time.Now()
	expected := append(v,v2...)
	if err2 != nil {
		t.Error("Failed to append and download", err2)
		return
	}
	if !reflect.DeepEqual(expected, vu2) {
		t.Error("Downloaded file is not the same", expected, vu2)
		return
	}

	// 3. long append test
	//ti =time.Now()
	v3 := []byte(", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 	 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 	 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 	 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 	 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 	 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 	 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 		 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 		 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 		 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 		 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 		 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 		 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 			 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 			 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 			 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 			 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 			 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 			 "scesidcgisavscb"+
 			 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 				 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 				 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 				 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 				 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 				 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 				 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 					 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 					 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 					 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 					 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 					 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 					 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
 						 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
 						 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
 						 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
 						 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
 						 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
 						 "scesidcgisavscb"+
						 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
							 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
							 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
							 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
							 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
							 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
							 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
								 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
								 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
								 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
								 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
								 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
								 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
									 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
									 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
									 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
									 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
									 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
									 "scesidcgisavscb"+
									 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
										 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
										 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
										 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
										 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
										 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
										 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
											 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
											 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
											 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
											 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
											 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
											 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
												 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
												 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
												 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
												 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
												 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
												 "scesidcgisavscb")

	u.AppendFile("file1", v3)
	//fmt.Println("long append complete", time.Since(ti))
	//ti =time.Now()
	vu3, err3 := u.LoadFile("file1")
	//fmt.Println("long load complete", time.Since(ti))
	//ti =time.Now()
	expected = append(expected,v3...)
	if err3 != nil {
		t.Error("Failed to append and download", err3)
		return
	}
	if !reflect.DeepEqual(expected, vu3) {
		t.Error("Downloaded file is not the same", expected, vu3)
		return
	}

	// 4. Nonexistence test
	err4 := u.AppendFile("file2", []byte("test"))
	if err4 == nil {
		t.Error("Failed non-existence file query test", err4)
		return
	}

	// 5. Second instantiation test
	u2, err5 := GetUser("a", "fubar")
	if err5 != nil {
		t.Error("Failed to support 2nd-instantiation of user", err5)
		return
	}

	// 6. Second user append and load himself
	v6 := []byte("This is another me")
	u2.AppendFile("file1", v6)
	vu6, err6 := u.LoadFile("file1")
	expected = append(expected,v6...)
	if err6 != nil {
		t.Error("Failed to append and download", err6)
		return
	}
	if !reflect.DeepEqual(expected, vu6) {
		t.Error("Downloaded file is not the same", expected, vu6)
		return
	}

	// 9. Append 100 times
	 u.StoreFile("file9", []byte(""))
	 temp9_v := "test9 "
	 expected9 := []byte("")
	 for i := 0; i < 20; i++ {
	  u2.AppendFile("file9", []byte(temp9_v + string(i)))
	  expected9 = append(expected9, []byte(temp9_v + string(i))...)
	 }
	 v9, err9 := u2.LoadFile("file9")
	 if err9 != nil {
	  t.Error("Failed when appedning 100 times", err9)
	  return
	 }
	 if !reflect.DeepEqual(expected9, v9) {
	  t.Error("Downloaded file is not the same for 100 times appending", expected9, v9)
	  return
	 }
	// destroy test
  clear()
	err7 := u2.AppendFile("file1", v6)
	if err7 == nil {
		t.Error("Got it after clear()", err7)
		return
	}
	return
}

func TestShare(t *testing.T) {
 clear()
 // simple share one-to-one test (share & receive)
 u, err := InitUser("a", "fubar")
 if err != nil {
  t.Error("Failed to initialize user", err)
  return
 }
 u2, err2 := InitUser("b", "foobar")
 if err2 != nil {
  t.Error("Failed to initialize b", err2)
  return
 }

 v := []byte("This is a test")
 u.StoreFile("file1", v)

 var v2 []byte
 var magic_string string

 v, err = u.LoadFile("file1")
 if err != nil {
  t.Error("Failed to download the file from a", err)
  return
 }

 //ti := time.Now()
 magic_string, err = u.ShareFile("file1", "b")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 //fmt.Println("short share complete", time.Since(ti))
 //ti =time.Now()
 err = u2.ReceiveFile("file2", "a", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 //fmt.Println("short receive complete", time.Since(ti))

 v2, err = u2.LoadFile("file2")
 if err != nil {
  t.Error("Failed to download the file after sharing", err)
  return
 }

 if !reflect.DeepEqual(v, v2) {
  t.Error("Shared file is not the same", v, v2)
  return
 }

 // 1-3-(4,5) 1-2-(6,7)
 u3, err3 := InitUser("c", "c")
 if err3 != nil {
  t.Error("Failed to initialize user", err3)
  return
 }
 //fmt.Println("share chain start")
 vp := []byte(", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
	 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
	 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
	 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
	 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
	 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
	 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
		 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
		 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
		 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
		 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
		 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
		 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
			 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
			 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
			 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
			 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
			 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
			 "scesidcgisavscb"+
			 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
				 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
				 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
				 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
				 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
				 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
				 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
					 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
					 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
					 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
					 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
					 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
					 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
						 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
						 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
						 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
						 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
						 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
						 "scesidcgisavscb"+", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
							 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
							 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
							 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
							 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
							 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
							 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
								 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
								 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
								 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
								 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
								 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
								 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
									 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
									 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
									 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
									 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
									 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
									 "scesidcgisavscb"+
									 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
										 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
										 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
										 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
										 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
										 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
										 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
											 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
											 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
											 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
											 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
											 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
											 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
												 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
												 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
												 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
												 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
												 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
												 "scesidcgisavscb")

 u.AppendFile("file1", vp)
 v = append(v, vp...)
 //ti =time.Now()
 magic_string, err = u.ShareFile("file1", "c")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 u.AppendFile("file1", []byte("funcy"))
 v = append(v, []byte("funcy")...)
 //fmt.Println("long share complete", time.Since(ti))
 //ti =time.Now()
 err = u3.ReceiveFile("file3", "a", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 //fmt.Println("long receive complete", time.Since(ti))

 v2, err = u3.LoadFile("file3")
 if err != nil {
  t.Error("Failed to download the file after sharing", err)
  return
 }

 if !reflect.DeepEqual(v, v2) {
  t.Error("Shared file is not the same", v, v2)
  return
 }

 u4, err4 := InitUser("d", "d")
 if err4 != nil {
  t.Error("Failed to initialize user", err4)
  return
 }
 magic_string4, err := u3.ShareFile("file3", "d")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }

 u5, err5 := InitUser("e", "e")
 if err5 != nil {
  t.Error("Failed to initialize user", err5)
  return
 }
 u5.StoreFile("file_5",[]byte("jweh"))
 magic_string, err = u3.ShareFile("file3", "e")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 err = u5.ReceiveFile("file_5", "c", magic_string)
 if err == nil {
 t.Error("wrong name already exists", err)
  return
 }
 err = u5.ReceiveFile("file5", "c", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }

 u6, err6 := InitUser("f", "f")
 if err6 != nil {
  t.Error("Failed to initialize user", err6)
  return
 }
 magic_string6, err := u2.ShareFile("file2", "f")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }


 u7, err7 := InitUser("g", "g")
 if err7 != nil {
  t.Error("Failed to initialize user", err7)
  return
 }
 magic_string, err = u2.ShareFile("file2", "g")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 err = u7.ReceiveFile("file7", "c", magic_string)
 if err == nil {
  t.Error("rip", err)
  return
 }
 err = u7.ReceiveFile("file7", "b", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 u8, err8 := InitUser("h", "foobar")
 if err8 != nil {
 t.Error("Failed to initialize h", err8)
 return
 }
 u9, err9 := InitUser("i", "foobar")
 if err9 != nil {
 t.Error("Failed to initialize i", err9)
 return
 }

 v89 := []byte("This is a test")
 u8.StoreFile("file8", v89)
 u9.StoreFile("file9",v89)
 m8, _ := u8.ShareFile("file8","b")
 m9, _ := u9.ShareFile("file9","b")
 err = u2.ReceiveFile("file8", "h", m8)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 err = u2.ReceiveFile("file9", "i", m9)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 m8, _ = u8.ShareFile("file8","c")
 m9, _ = u9.ShareFile("file9","c")
 err = u3.ReceiveFile("file8", "h", m8)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 err = u3.ReceiveFile("file9", "i", m9)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 m8, _ = u2.ShareFile("file8","g")
 m9, _ = u3.ShareFile("file9","d")
 err = u7.ReceiveFile("file8", "b", m8)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 m48, _ := u3.ShareFile("file8","d")
 //m9, _ = u3.ShareFile("file9","d")


 //ti =time.Now()
 ////fmt.Println("revoking b...")

 err = u.RevokeFile("file1", "f")
 if err == nil {
  t.Error("revoked a non direct child", err)
  return
 }
 v2, err = u7.LoadFile("file7")
 if !reflect.DeepEqual(v, v2) {
	t.Error("Failed to revoke 7", v, v2)
	return
 }

 err = u.RevokeFile("file1", "b")
 if err != nil {
  t.Error("Failed to revoke the file", err)
  return
 }
 err = u8.RevokeFile("file8", "b")
 if err != nil {
  t.Error("Failed to revoke the file", err)
  return
 }
 err = u9.RevokeFile("file9", "b")
 if err != nil {
  t.Error("Failed to revoke the file", err)
  return
 }

 v_89 := []byte(", kid you not")
 u8.StoreFile("file8", v_89)
 u9.StoreFile("file9", v_89)
 v89 = v_89

 //fmt.Println("revoke complete", time.Since(ti))

 v_st := []byte(", kid you not")
 u.StoreFile("file1", v_st)
 v = v_st
 //ti =time.Now()
 v_ap := []byte(", kid you not")
 err = u3.AppendFile("file3", v_ap)
 if err != nil {
  t.Error("failed append", err)
  return
 }
 v = append(v, v_ap...)

 //fmt.Println("append after revoke complete", time.Since(ti))

 v2, err = u3.LoadFile("file3")
 if !reflect.DeepEqual(v, v2) {
	t.Error("revoked 3 by accident", v, v2)
	return
 }
 v2, err = u.LoadFile("file1")
 if !reflect.DeepEqual(v, v2) {
  t.Error("user 1 has bad file", v, v2)
  return
 }

v2, err = u3.LoadFile("file8")
 if !reflect.DeepEqual(v89, v2) {
  t.Error("revoked 3 by accident", v89, v2)
  return
 }

 v2, err = u2.LoadFile("file2")
 if reflect.DeepEqual(v, v2) {
  t.Error("Failed to revoke 2", v, v2)
  return
 }
 v2, err = u2.LoadFile("file9")
 if reflect.DeepEqual(v89, v2) {
  t.Error("revoked 3 by accident", v89, v2)
  return
 }
 err = u4.ReceiveFile("file4", "c", magic_string4)
 if err != nil {
	t.Error("Failed to receive the share message", err)
	return
 }
 v2, err = u4.LoadFile("file4")
 if !reflect.DeepEqual(v, v2) {
  t.Error("revoked 4 by accident", string(v), string(v2))
  return
 }
 err = u4.ReceiveFile("file9", "c", m9)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 v2, err = u4.LoadFile("file9")
 if !reflect.DeepEqual(v89, v2) {
  t.Error("revoked 3 by accident", string(v89), string(v2))
  return
 }
 err = u4.ReceiveFile("file8", "c", m48)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 v2, err = u4.LoadFile("file8")
 if !reflect.DeepEqual(v89, v2) {
  t.Error("revoked 3 by accident", v89, v2)
  return
 }

 v2, err = u5.LoadFile("file5")
 if !reflect.DeepEqual(v, v2) {
  t.Error("revoked 5 by accident", v, v2)
  return
 }


 err = u6.ReceiveFile("file6", "b", "ole")
 if err == nil {
	t.Error("Failed to receive the share message", err)
	return
 }
 err = u6.ReceiveFile("file6", "b", magic_string6)
 if err != nil {
	t.Error("Failed to receive the share message", err)
	return
 }

 v2, err = u6.LoadFile("file6")
 if reflect.DeepEqual(v, v2) {
  t.Error("Failed to revoke 6", v, v2)
  return
 }


 v2, err = u7.LoadFile("file7")
 if reflect.DeepEqual(v, v2) {
  t.Error("Failed to revoke 7", v, v2)
  return
 }
 v2, err = u7.LoadFile("file8")
 if reflect.DeepEqual(v89, v2) {
  t.Error("revoked 3 by accident", v89, v2)
  return
 }

 // 9: Try to share to a nonexist user
 _, err = u2.ShareFile("file2", "nonexist")
 if err == nil {
  t.Error("Error when share to a nonexist user", err)
  return
 }

 _, err = u2.ShareFile("filechk", "f")
 if err == nil {
  t.Error("Error when sharing a nonexist file", err)
  return
 }
 return
}

func TestShare2(t *testing.T) {
 clear()
 // simple share one-to-one test (share & receive)
 u, err := InitUser("a", "fubar")
 if err != nil {
  t.Error("Failed to initialize user", err)
  return
 }
 u2, err2 := InitUser("b", "foobar")
 if err2 != nil {
  t.Error("Failed to initialize b", err2)
  return
 }

 v := []byte("This is a test")
 u.StoreFile("file1", v)

 var v2 []byte
 var magic_string string

 v, err = u.LoadFile("file1")
 if err != nil {
  t.Error("Failed to download the file from a", err)
  return
 }

 //ti := time.Now()
 magic_string, err = u.ShareFile("file1", "b")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 //fmt.Println("short share complete", time.Since(ti))
 //ti =time.Now()
 err = u2.ReceiveFile("file2", "a", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 //fmt.Println("short receive complete", time.Since(ti))

 v2, err = u2.LoadFile("file2")
 if err != nil {
  t.Error("Failed to download the file after sharing", err)
  return
 }

 if !reflect.DeepEqual(v, v2) {
  t.Error("Shared file is not the same", v, v2)
  return
 }

 // 1-3-(4,5) 1-2-(6,7)
 u3, err3 := InitUser("c", "c")
 if err3 != nil {
  t.Error("Failed to initialize user", err3)
  return
 }
 //fmt.Println("share chain start")
 vp := []byte(", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
	 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
	 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
	 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
	 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
	 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
	 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
		 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
		 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
		 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
		 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
		 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
		 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
			 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
			 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
			 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
			 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
			 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
			 "scesidcgisavscb"+
			 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
				 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
				 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
				 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
				 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
				 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
				 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
					 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
					 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
					 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
					 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
					 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
					 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
						 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
						 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
						 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
						 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
						 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
						 "scesidcgisavscb"+", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
							 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
							 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
							 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
							 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
							 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
							 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
								 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
								 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
								 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
								 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
								 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
								 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
									 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
									 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
									 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
									 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
									 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
									 "scesidcgisavscb"+
									 ", kid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
										 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
										 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
										 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
										 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
										 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
										 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
											 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
											 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
											 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
											 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
											 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
											 "scesidcgisavscbkid you notvdqhhhhhhhhhhshverisydvceiwahrsdkcarsdiuyjhxc"+
												 "cbahjriscviweruagsvdvehrsdfkbvegoruigslviab fhsdvfiasduicbaksdfouveqra"+
												 "cbjvhrasviuxgcorueqgwsofduvcihrsdbouvewahklqfavw;ifguv;rpdgisuflvraegwsduf"+
												 "ajhcbvsidvxicgviejsduhcjixewgasvdguvcwjegsadsuifcvwvyiegascvxhwevsadivc"+
												 "wcevdgjvc weaiusdhcgviwegasbvfcigaebsvfieguasbfvwiegasfvcisadufcbvweisagdc"+
												 "cveisugdcvixgweoabvficgesabvdciugweabvdfigcuowhaedfvcgouaehfvgoucaefweasfae"+
												 "scesidcgisavscb")

 u.AppendFile("file1", vp)
 v = append(v, vp...)
 //ti =time.Now()
 magic_string, err = u.ShareFile("file1", "c")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 u.AppendFile("file1", []byte("funcy"))
 v = append(v, []byte("funcy")...)
 //fmt.Println("long share complete", time.Since(ti))
 //ti =time.Now()
 err = u3.ReceiveFile("file3", "a", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }
 //fmt.Println("long receive complete", time.Since(ti))

 v2, err = u3.LoadFile("file3")
 if err != nil {
  t.Error("Failed to download the file after sharing", err)
  return
 }

 if !reflect.DeepEqual(v, v2) {
  t.Error("Shared file is not the same", v, v2)
  return
 }

 u4, err4 := InitUser("d", "d")
 if err4 != nil {
  t.Error("Failed to initialize user", err4)
  return
 }
 magic_string, err = u3.ShareFile("file3", "d")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 err = u4.ReceiveFile("file4", "c", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }

 u5, err5 := InitUser("e", "e")
 if err5 != nil {
  t.Error("Failed to initialize user", err5)
  return
 }
 u5.StoreFile("file_5",[]byte("jweh"))
 magic_string, err = u3.ShareFile("file3", "e")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 err = u5.ReceiveFile("file_5", "c", magic_string)
 if err == nil {
 t.Error("wrong name already exists", err)
  return
 }
 err = u5.ReceiveFile("file5", "c", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }

 u6, err6 := InitUser("f", "f")
 if err6 != nil {
  t.Error("Failed to initialize user", err6)
  return
 }
 magic_string6, err := u2.ShareFile("file2", "f")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }


 u7, err7 := InitUser("g", "g")
 if err7 != nil {
  t.Error("Failed to initialize user", err7)
  return
 }
 magic_string, err = u2.ShareFile("file2", "g")
 if err != nil {
  t.Error("Failed to share the a file", err)
  return
 }
 err = u7.ReceiveFile("file7", "c", magic_string)
 if err == nil {
  t.Error("rip", err)
  return
 }
 err = u7.ReceiveFile("file7", "b", magic_string)
 if err != nil {
  t.Error("Failed to receive the share message", err)
  return
 }

 //ti =time.Now()
 ////fmt.Println("revoking b...")
 /*
 err = u.RevokeFile("file1", "f")
 if err == nil {
  t.Error("revoked a non direct child", err)
  return
 }
 v2, err = u7.LoadFile("file7")
 if !reflect.DeepEqual(v, v2) {
	t.Error("Failed to revoke 7", v, v2)
	return
 }*/
 err = u.RevokeFile("file1", "b")
 if err != nil {
  t.Error("Failed to revoke the file", err)
  return
 }
 //fmt.Println("revoke complete", time.Since(ti))

 v_st := []byte(", kid you not")
 u.StoreFile("file1", v_st)
 v = v_st
 //ti =time.Now()
 v_ap := []byte(", kid you not")
 u3.AppendFile("file3", v_ap)
 v = append(v, v_ap...)
 //fmt.Println("append after revoke complete", time.Since(ti))

 v2, err = u.LoadFile("file1")
 if !reflect.DeepEqual(v, v2) {
  t.Error("user 1 has bad file", v, v2)
  return
 }

 v2, err = u3.LoadFile("file3")
 if !reflect.DeepEqual(v, v2) {
  t.Error("revoked 3 by accident", v, v2)
  return
 }

 v2, err = u2.LoadFile("file2")
 if reflect.DeepEqual(v, v2) {
  t.Error("Failed to revoke 2", v, v2)
  return
 }

 v2, err = u.LoadFile("file1")
 if !reflect.DeepEqual(v, v2) {
  t.Error("user 1 has bad file", v, v2)
  return
 }

 v2, err = u3.LoadFile("file3")
 if !reflect.DeepEqual(v, v2) {
  t.Error("revoked 3 by accident", v, v2)
  return
 }

 v2, err = u4.LoadFile("file4")
 if !reflect.DeepEqual(v, v2) {
  t.Error("revoked 4 by accident", v, v2)
  return
 }

 v2, err = u5.LoadFile("file5")
 if !reflect.DeepEqual(v, v2) {
  t.Error("revoked 5 by accident", v, v2)
  return
 }

 err = u6.ReceiveFile("file6", "b", magic_string)
 if err == nil {
	t.Error("Failed to receive the share message", err)
	return
 }
 err = u6.ReceiveFile("file6", "b", magic_string6)
 if err != nil {
	t.Error("Failed to receive the share message", err)
	return
 }

 v2, err = u6.LoadFile("file6")
 if reflect.DeepEqual(v, v2) {
  t.Error("Failed to revoke 6", v, v2)
  return
 }


 v2, err = u7.LoadFile("file7")
 if reflect.DeepEqual(v, v2) {
  t.Error("Failed to revoke 7", v, v2)
  return
 }

 // 9: Try to share to a nonexist user
 _, err = u2.ShareFile("file2", "nonexist")
 if err == nil {
  t.Error("Error when share to a nonexist user", err)
  return
 }

 _, err = u2.ShareFile("filechk", "f")
 if err == nil {
  t.Error("Error when sharing a nonexist file", err)
  return
 }
 return
}
