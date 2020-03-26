package proj2

// CS 161 Project 2 Spring 2020
// You MUST NOT change what you import.  If you add ANY additional
// imports it will break the autograder. We will be very upset.

import (
	//"fmt"
	//"time"
	"github.com/cs161-staff/userlib"
	"encoding/json"
	//"encoding/hex"
	"github.com/google/uuid"
	"strings"
	"errors"
	_ "strconv"
)

func someUsefulThings() {
	// Creates a random UUID
	//fmt.Println("dummy")
}
// TODO: write first then decide if you need creator (no need if no test for it)
// TODO: combine syncuser and SyncFile
// TODO: redundancy check
// TODO: fix the verifying jacket with after unmarshall problem
// TODO: C is not recursing down because it can't keep the same file name
// *** Beginning of Data Format *** //
/*
	DS/DV/SK/PK: 16
	SymKey: 16
	Signature: 256
	[]byte UUID: 16
	string UUID: 36
	HMACEval: 64
	Jacket: salt (16) | hashed&salted usrname+pswd (64) | <user-sign-key>-signature of jacket serial (256) | jacket serial (<pswd-gen-symkey>-encrypted user struct serial + unencrypted fields) (...)
	File: <file-sign-key>-signature of file struct serial (256) | file struct serial (<file-symKey>-encrypted filecontent) (...)
	Magic string: <sender-sign-key>-signature of sender+encrypted access_token (256) | <recipient-public-key>-encrypted access_token (...)
	Access token: one-time use unlock-key (16) | string recipient_shell_struct_uuid (36)
*/
// *** End of Datastore Format *** //
// *** Beginning of Structure Definitions *** //
type User struct {
	Username string
	Password string
	Dict_to_file_struct map[string]uuid.UUID
	Dict_to_original_file_struct map[string]uuid.UUID
	Dict_to_file_decrypt map[string][]byte
	Dict_to_file_verify map[string]userlib.DSVerifyKey
	SK userlib.PKEDecKey
	PK userlib.PKEEncKey
	EncLong []byte
	Sign userlib.DSSignKey
}
type Jacket struct {
	User []byte
	UpdatedAddress map[uuid.UUID][]byte //
	LastCreator string // multiple revokes may affect you
	Receiver_dict map[uuid.UUID][]string //username->filename
	Direct_Children map[string]bool
}
type File struct {
	Addresses []uuid.UUID
	SignKey userlib.DSSignKey
}

// *** End of Structure Definitions *** //
// *** Beginning of Helper Definitions *** //
func (userdata *User) SyncUser() (jacket *Jacket, queried_jacket []byte, e error){
	user_byte_id, err := userlib.HMACEval(make([]byte, 16),[]byte(userdata.Username))
	if err != nil {
		return nil, nil, err
	}
	user_uuid, err := uuid.FromBytes(user_byte_id[:16])
	if err != nil {
		return nil, nil, err
	}
	queried_jacket, ok := userlib.DatastoreGet(user_uuid)
	if ok == false {
		return nil, nil, errors.New(strings.ToTitle("Cannot fetch from datastore"))
	}
	if len(queried_jacket) < 336 {
		return nil, nil, errors.New(strings.ToTitle("Jacket struct broken"))
	}

	sus_signature := queried_jacket[80:336]
	jacket_ser := queried_jacket[336:]
	var j Jacket
	jacket = &j
	json.Unmarshal(queried_jacket[336:], jacket)

	usr_verify_key, e := GetVerifyKey(userdata.Username)
	if e != nil {
    return nil, nil, e
  }
	error_user := userlib.DSVerify(usr_verify_key, jacket_ser, sus_signature)
	if len(jacket.LastCreator) != 0 {
		creator_verify_key, e := GetVerifyKey(jacket.LastCreator)
		if e != nil {
	    return nil, nil, e
	  }
		error_creator := userlib.DSVerify(creator_verify_key, jacket_ser, sus_signature)
		if !(error_user == nil || error_creator == nil)  {
			return nil, nil, errors.New(strings.ToTitle("Jacket struct integrity has been compromised"))
		}
	} else {
		if error_user != nil {
			return nil, nil, errors.New(strings.ToTitle("Jacket struct integrity has been compromised"))
		}
	}

	user_dec_key := userlib.Argon2Key([]byte(userdata.Password),[]byte("0"),16)
	dec_curr_user := userlib.SymDec(user_dec_key, jacket.User)
	err = json.Unmarshal(dec_curr_user, userdata)
	if err != nil {
    return nil, nil, errors.New(strings.ToTitle("Can't deserialize"))
  }
	return jacket, queried_jacket, nil
}

func (userdata *User) SyncFile() error {
	 jacket, queried_jacket, e := userdata.SyncUser()
	 if e != nil {
	   return e
	 }

	 if (jacket.UpdatedAddress != nil) {
		tmp := make(map[uuid.UUID]string)
    for k, v := range userdata.Dict_to_file_struct {
        tmp[v] = k
    }
		 for k, v := range jacket.UpdatedAddress {
			 	dec_updated_addr, e := userlib.PKEDec(userdata.SK, v)
				dec_updated_addr = dec_updated_addr[36:] //this is the new address
	 			if e != nil {
	 		    return e
	 		  }
				new_loc, e := uuid.Parse(string(dec_updated_addr))
				if e != nil {
			    return e
			  }
				_, ok := tmp[k]
				if ok == true {
					name_for_uuid := tmp[k]
					userdata.Dict_to_file_struct[name_for_uuid] = new_loc
					delete(jacket.UpdatedAddress, k)
				}
			 }
		 if len(queried_jacket) < 80 {
			 return errors.New(strings.ToTitle("Jacket struct broken"))
		 }
		 e = userdata.UpdateUser(jacket, queried_jacket[:80])
		 if e != nil {
			 return e
		 }
	 }
	 return nil
}
func (userdata *User) SyncFileForReceive(filename string, original_shared_addr string) error {
	 jacket, queried_jacket, e := userdata.SyncUser()
	 if e != nil {
	   return e
	 }
	 if (jacket.UpdatedAddress != nil) {
		tmp := make(map[uuid.UUID]string)
    for k, v := range userdata.Dict_to_file_struct {
        tmp[v] = k
    }
		 for k, v := range jacket.UpdatedAddress {
			 	dec_updated_addr, e := userlib.PKEDec(userdata.SK, v)
				dec_target_old := string(dec_updated_addr[:36])
				dec_target_new := string(dec_updated_addr[36:])
	 			if e != nil {
	 		    return e
	 		  }
				new_loc, e := uuid.Parse(dec_target_new)
				if e != nil {
			    return e
			  }
				_, ok := tmp[k]
				if ok == true {
				    name_for_uuid := tmp[k]
						userdata.Dict_to_file_struct[name_for_uuid] = new_loc
						delete(jacket.UpdatedAddress, k)
				} else {
						if original_shared_addr == dec_target_old {
						  name_for_uuid := filename
							userdata.Dict_to_file_struct[name_for_uuid] = new_loc
							delete(jacket.UpdatedAddress, k)
						}
				}
			 }
		 //jacket.UpdatedAddress = nil
		 if len(queried_jacket) < 80 {
			 return errors.New(strings.ToTitle("Jacket struct broken"))
		 }
		 e = userdata.UpdateUser(jacket, queried_jacket[:80])
		 if e != nil {
			 return e
		 }
	 }
	 return nil
}
func GetUserUUID(username string) (id uuid.UUID) {
	user_byte_id, _ := userlib.HMACEval(make([]byte, 16),[]byte(username))
	user_uuid, _ := uuid.FromBytes(user_byte_id[:16])
	return user_uuid
}
func (userdata *User) UpdateUser(jacket *Jacket, prefix []byte) error {
	user_byte_id, e := userlib.HMACEval(make([]byte, 16),[]byte(userdata.Username))
	if e != nil {
		return e
	}
	user_uuid, e := uuid.FromBytes(user_byte_id[:16])
	if e != nil {
		return e
	}
	user_sym_key := userlib.Argon2Key([]byte(userdata.Password),[]byte("0"),16)
	user_packaged_data, e := json.Marshal(userdata)
	if e != nil {
		return e
	}
	encrypted_user := userlib.SymEnc(user_sym_key, userlib.RandomBytes(16), user_packaged_data)
	jacket.User = encrypted_user
	// Serializing jacket struct
	jacket_ser, e := json.Marshal(jacket)
	if e != nil {
		return e
	}
	// Sign the jacket serial
	signature, e := userlib.DSSign(userdata.Sign, jacket_ser)
	if e != nil {
		return e
	}
	to_upload  := append(append(prefix, signature...), jacket_ser...)
	userlib.DatastoreSet(user_uuid, to_upload)
	return nil
}
func GetVerifyKey(username string) (vk userlib.DSVerifyKey, err error) {
	verify_key_id, e := userlib.HMACEval(make([]byte, 16),[]byte(username+"verify"))
	_, dummy, e := userlib.DSKeyGen()
	if e != nil {
		return dummy, e
	}
	verify_key, ok := userlib.KeystoreGet(string(verify_key_id))
	if ok == false {
		return dummy, errors.New(strings.ToTitle("Could not fetch verify key for "+username))
	}
	return verify_key, nil
}
// *** End of Helper Definitions *** //
// *** Beginning of Interface Implementations *** //
/*
	// This creates a user.  It will only be called once for a user
	// (unless the keystore and datastore are cleared during testing purposes)

	// It should store a copy of the userdata, suitably encrypted, in the
	// datastore and should store the user's public key in the keystore.

	// The datastore may corrupt or completely erase the stored
	// information, but nobody outside should be able to get at the stored

	// You are not allowed to use any global storage other than the
	// keystore and the datastore functions in the userlib library.

	// You can assume the password has strong entropy, EXCEPT
	// the attackers may possess a precomputed tables containing
	// hashes of common passwords downloaded from the internet.
*/


func InitUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

	// Check for existing user
	user_uuid := GetUserUUID(username)
	_, ok := userlib.DatastoreGet(user_uuid)
	if ok == true {
		return nil, errors.New(strings.ToTitle("User already exists"))
	}

	// Field Population
	userdata.Username = username
	userdata.Password = password
	userdata.Dict_to_file_struct = make(map[string]uuid.UUID)
	userdata.Dict_to_file_decrypt = make(map[string][]byte)
	userdata.Dict_to_file_verify = make(map[string]userlib.DSVerifyKey)
	userdata.Dict_to_original_file_struct = make(map[string]uuid.UUID)
	pk, sk, e := userlib.PKEKeyGen()
	if e != nil {
		return nil, e
	}
	sign, verify, e := userlib.DSKeyGen()
	if e != nil {
		return nil, e
	}
	userdata.SK = sk
	userdata.PK = pk
	userdata.Sign = sign
	user_sym_key := userlib.Argon2Key([]byte(password),[]byte("0"),16)
	user_enc_long, e := userlib.HashKDF(user_sym_key, []byte("enc_long"))
	if e != nil {
		return nil, e
	}
 	user_enc_long = user_enc_long[:16]
 	userdata.EncLong = user_enc_long

	// Storing PK and Verifying key in Keystore
	public_key_id, e := userlib.HMACEval(make([]byte, 16),[]byte(username+"public"))
	if e != nil {
		return nil, e
	}
	userlib.KeystoreSet(string(public_key_id), pk)
	verify_key_id, e := userlib.HMACEval(make([]byte, 16),[]byte(username+"verify"))
	if e != nil {
		return nil, e
	}
	userlib.KeystoreSet(string(verify_key_id), verify)

	// Serializing & Encrypting userdata
	ser, e := json.Marshal(userdata)
	if e != nil {
		return nil, e
	}
	encrypted_ser := userlib.SymEnc(user_sym_key, userlib.RandomBytes(16), ser)

	// Create Jacket Struct
	var j Jacket
	jacket := &j
	j.User = encrypted_ser
	j.LastCreator = ""
	j.UpdatedAddress = make(map[uuid.UUID][]byte)
	j.Receiver_dict = make(map[uuid.UUID][]string)
	j.Direct_Children = make(map[string]bool)

	// Serializing jacket struct
	jacket_ser, e := json.Marshal(jacket)
	if e != nil {
		return nil, e
	}

	// Sign the jacket serial
	signature, e := userlib.DSSign(sign, jacket_ser)
	if e != nil {
		return nil, e
	}

	// Uploading to Datastore with correct prefix
	salt := userlib.RandomBytes(16)
	pswd_hash, e := userlib.HMACEval(salt, []byte(username+password)) //salted since attacker can dictionary attack
	if e != nil {
		return nil, e
	}
	to_upload := append(append(append(salt, pswd_hash...), signature...), jacket_ser...)
	userlib.DatastoreSet(user_uuid, to_upload)
	return userdataptr, nil
}
/*
	// This fetches the user information from the Datastore.  It should
	// fail with an error if the user/password is invalid, or if the user
	// data was corrupted, or if the user can't be found.
*/

func GetUser(username string, password string) (userdataptr *User, err error) {
	var userdata User
	userdataptr = &userdata

	// Verify hashed salted password
	user_uuid := GetUserUUID(username)
	queried_jacket, ok := userlib.DatastoreGet(user_uuid)
	if len(queried_jacket) < 336 {
		return nil, errors.New(strings.ToTitle("Jacket struct broken"))
	}
	if ok == false {
		return nil, errors.New(strings.ToTitle("Unable to fetch jacket for "+username))
	}
	salt := queried_jacket[:16]
	correct_hash := queried_jacket[16:80]
	sus_has, e := userlib.HMACEval(salt, []byte(username+password))
	if e != nil {
		return nil, e
	}
	if userlib.HMACEqual(sus_has, correct_hash) == false {
		return nil, errors.New(strings.ToTitle("Incorrect username or password"))
	}

	// Unmarshall jacket
	sus_signature := queried_jacket[80:336]
	jacket_ser := queried_jacket[336:]
	var j Jacket
	jacket := &j
	json.Unmarshal(jacket_ser, jacket)

	// Verify integrity of jacket
	usr_verify_key, e := GetVerifyKey(username)
	if e != nil {
    return nil, e
  }
	error_user := userlib.DSVerify(usr_verify_key, jacket_ser, sus_signature)
	if len(jacket.LastCreator) != 0 {
		creator_verify_key, e := GetVerifyKey(jacket.LastCreator)
		if e != nil {
	    return nil, e
	  }
		error_creator := userlib.DSVerify(creator_verify_key, jacket_ser, sus_signature)
		if !(error_user == nil || error_creator == nil)  {
			return nil, errors.New(strings.ToTitle("Jacket struct integrity has been compromised"))
		}
	} else {
		if error_user != nil {
			return nil, errors.New(strings.ToTitle("Jacket struct integrity has been compromised"))
		}
	}

	// Unmarshall & Decrypt user
	encrypted_ser := jacket.User
	user_sym_key := userlib.Argon2Key([]byte(password),[]byte("0"),16)
	ser := userlib.SymDec(user_sym_key, encrypted_ser)
	json.Unmarshal(ser, userdataptr)

	return userdataptr, nil
}
/*
	// This stores a file in the datastore.
	//
	// The plaintext of the filename + the plaintext and length of the filename
	// should NOT be revealed to the datastore!
*/
func (userdata *User) StoreFile(filename string, data []byte) {
	 // Check integrity and sync with the current userstruct
	 e := userdata.SyncFile()
	 if e != nil {
		 userlib.DebugMsg(strings.ToTitle(e.Error()))
		 return
	 }
	 jacket, queried_jacket, e := userdata.SyncUser()
	 if e != nil {
		 userlib.DebugMsg(strings.ToTitle(e.Error()))
		 return
	 }
	 _, exist := userdata.Dict_to_file_struct[filename]
	 if exist == true {
		 // We leave shell and share unchanged because the sharing relationship persists
		 uuid_newfileContent := uuid.New()
		 queried_file, ok := userlib.DatastoreGet(userdata.Dict_to_file_struct[filename])
		 if ok == false {
			 return
		 }
		 if len(queried_file) < 256 {
			 return
		 }

		 error_file := userlib.DSVerify(userdata.Dict_to_file_verify[filename], queried_file[256:], queried_file[:256])
		 if error_file != nil {
			 return
		 }

		 dec_queried_file := userlib.SymDec(userdata.Dict_to_file_decrypt[filename], queried_file[256:])
		 var file File
		 file_struct := &file
		 e = json.Unmarshal(dec_queried_file, file_struct)
		 if e != nil {
			 return
		 }
		 // Reset starting addresses only, sign key remains
		 file_struct.Addresses = append([]uuid.UUID{}, uuid_newfileContent)
		 // Update this file struct
		 updated_file_packaged_data, _ := json.Marshal(file_struct)
		 enc_updated_file_packaged_data := userlib.SymEnc(userdata.Dict_to_file_decrypt[filename], userlib.RandomBytes(16), updated_file_packaged_data)
		 updated_fileStruct_sign, _ := userlib.DSSign(file_struct.SignKey, enc_updated_file_packaged_data)
		 updated_fileStruct_datastore := append(updated_fileStruct_sign, enc_updated_file_packaged_data...)
		 userlib.DatastoreSet(userdata.Dict_to_file_struct[filename], updated_fileStruct_datastore)

		 // Store the file content at uuid_fileContent
		 newenc_fileContent := userlib.SymEnc(userdata.Dict_to_file_decrypt[filename], userlib.RandomBytes(16), data)
		 newfileContent_sign, _ := userlib.DSSign(file_struct.SignKey, newenc_fileContent)
		 newfileContent_datastore := append(newfileContent_sign, newenc_fileContent...)
		 userlib.DatastoreSet(uuid_newfileContent, newfileContent_datastore)

		 return
	 }

	 // Initiation
	 uuid_fileStruct := uuid.New()
	 uuid_fileContent := uuid.New()
	 symKey := userlib.Argon2Key([]byte(userdata.Password), userlib.RandomBytes(16), 16)
	 filesignKey, fileVerifykey, _ := userlib.DSKeyGen()

	 // Creating the shareStruct
	 userdata.Dict_to_file_struct[filename] = uuid_fileStruct
	 userdata.Dict_to_original_file_struct[filename] = uuid_fileStruct
	 userdata.Dict_to_file_verify[filename] = fileVerifykey
	 userdata.Dict_to_file_decrypt[filename] = symKey
	 if len(queried_jacket) < 80 {
		 userlib.DebugMsg(strings.ToTitle("Jacket struct broken"))
	 }
	 e = userdata.UpdateUser(jacket, queried_jacket[:80])
	 if e != nil {
		 return
	 }

	 // Creating the fileStruct
	 var fileStruct File
	 fileStruct.Addresses = append([]uuid.UUID{}, uuid_fileContent)
	 fileStruct.SignKey = filesignKey

	 // Store the file content at uuid_fileContent
	 enc_fileContent := userlib.SymEnc(symKey, userlib.RandomBytes(16), data)
	 fileContent_sign, _ := userlib.DSSign(filesignKey, enc_fileContent)
	 fileContent_datastore := append(fileContent_sign, enc_fileContent...)
	 userlib.DatastoreSet(uuid_fileContent, fileContent_datastore)

	 // Store the file into the datastore
	 file_packaged_data, _ := json.Marshal(fileStruct)
	 enc_file_packaged_data := userlib.SymEnc(symKey, userlib.RandomBytes(16), file_packaged_data)
	 fileStruct_sign, _ := userlib.DSSign(filesignKey, enc_file_packaged_data)
	 fileStruct_datastore := append(fileStruct_sign, enc_file_packaged_data...)
	 userlib.DatastoreSet(uuid_fileStruct, fileStruct_datastore)

	 return
}

/*
	// This adds on to an existing file.
	//
	// Append should be efficient, you shouldn't rewrite or reencrypt the
	// existing file, but only whatever additional information and
	// metadata you need.
*/

func (userdata *User) AppendFile(filename string, data []byte) (err error) {
	// Check integrity and sync with the current userstruct
	e := userdata.SyncFile()
	if e != nil {
		return e
	}
	_, _, e = userdata.SyncUser()
	if e != nil {
		return e
	}

	file_uuid := userdata.Dict_to_file_struct[filename]
	//fmt.Println(userdata.Username, file_uuid)
	file_verify := userdata.Dict_to_file_verify[filename]

	// Verify the integrity of the file
	queried_file, ok := userlib.DatastoreGet(file_uuid)
	if ok == false {
		return errors.New(strings.ToTitle("Cannot fetch from datastore"))
	}
	if len(queried_file) < 256 {
		return errors.New(strings.ToTitle("File struct broken"))
	}
	error_file := userlib.DSVerify(file_verify, queried_file[256:], queried_file[:256])
	if error_file != nil {
		return errors.New(strings.ToTitle("File integrity has been compromised"))
	}

	// Update next_addresses
	uuidNF := uuid.New()
	dec_queried_file := userlib.SymDec(userdata.Dict_to_file_decrypt[filename], queried_file[256:])
	var file File
	file_struct := &file
	e = json.Unmarshal(dec_queried_file, file_struct)
	if e != nil {
		return errors.New(strings.ToTitle("Can't deserialize"))
	}
	next_addresses := file_struct.Addresses
	file_struct.Addresses = append(next_addresses, uuidNF)

	// Update this file
	updated_file_packaged_data, _ := json.Marshal(file_struct)
	enc_updated_file_packaged_data := userlib.SymEnc(userdata.Dict_to_file_decrypt[filename], userlib.RandomBytes(16), updated_file_packaged_data)
	updated_fileStruct_sign, _ := userlib.DSSign(file_struct.SignKey, enc_updated_file_packaged_data)
	updated_fileStruct_datastore := append(updated_fileStruct_sign, enc_updated_file_packaged_data...)
	userlib.DatastoreSet(userdata.Dict_to_file_struct[filename], updated_fileStruct_datastore)

	// Store next file
	enc_nextfileContent := userlib.SymEnc(userdata.Dict_to_file_decrypt[filename], userlib.RandomBytes(16), data)
	nextfileContent_sign, _ := userlib.DSSign(file_struct.SignKey, enc_nextfileContent)
	nextfileContent_datastore := append(nextfileContent_sign, enc_nextfileContent...)
	userlib.DatastoreSet(uuidNF, nextfileContent_datastore)

	return nil
}
/*
	// This loads an existing file.
*/
func (userdata *User) LoadFile(filename string) (data []byte, err error) {
	// Check integrity and sync with the current userstruct
	e := userdata.SyncFile()
	if e != nil {
		return nil, e
	}
	_, _, e = userdata.SyncUser()
	if e != nil {
		return nil, e
	}

  file_uuid := userdata.Dict_to_file_struct[filename]
	file_verify := userdata.Dict_to_file_verify[filename]

	// Verify the integrity of the file
	queried_file, ok := userlib.DatastoreGet(file_uuid)
	if ok == false {
		return nil, errors.New(strings.ToTitle("Cannot fetch from datastore"))
	}
	if len(queried_file) < 256 {
		return nil, errors.New(strings.ToTitle("File struct broken"))
	}
  error_file := userlib.DSVerify(file_verify, queried_file[256:], queried_file[:256])
	if error_file != nil {
	  return nil, errors.New(strings.ToTitle("File integrity has been compromised"))
	}

	// Get content of the plaintext
	dec_queried_file := userlib.SymDec(userdata.Dict_to_file_decrypt[filename], queried_file[256:])
	var file File
	file_struct := &file
	e = json.Unmarshal(dec_queried_file, file_struct)
	if e != nil {
    return nil, errors.New(strings.ToTitle("Can't deserialize"))
  }

	r_val := []byte{}
	for i := 0; i < len(file_struct.Addresses); i++ {
		temp, er := getContentFromAddress(file_verify, userdata.Dict_to_file_decrypt[filename], file_struct.Addresses[i])
		if er != nil {
		  return nil, er
		}
		r_val = append(r_val, temp...)
  }
	return r_val, nil
}

func getContentFromAddress(verifyKey userlib.DSVerifyKey, decryptKey []byte, address uuid.UUID) (r []byte, err error) {
	queried_file, ok := userlib.DatastoreGet(address)
	if ok == false {
		return nil, errors.New(strings.ToTitle("Cannot fetch from datastore"))
	}
	if len(queried_file) < 256 {
		return nil, errors.New(strings.ToTitle("File struct broken"))
	}
	error_file := userlib.DSVerify(verifyKey, queried_file[256:], queried_file[:256])
	if error_file != nil {
	  return nil, errors.New(strings.ToTitle("File integrity has been compromised"))
	}
	return userlib.SymDec(decryptKey, queried_file[256:]), nil
}

/*
	// This creates a sharing record, which is a key pointing to something
	// in the datastore to share with the recipient.

	// This enables the recipient to access the encrypted file as well
	// for reading/appending.

	// Note that neither the recipient NOR the datastore should gain any
	// information about what the sender calls the file.  Only the
	// recipient can access the sharing record, and only the recipient
	// should be able to know the sender.
*/

func (userdata *User) ShareFile(filename string, recipient string) (
	magic_string string, err error) {

	// Check integrity and sync with the current userstruct
	e := userdata.SyncFile()
	if e != nil {
		return "", e
	}
	jacket, queried_jacket, e := userdata.SyncUser()
	if e != nil {
		return "", e
	}

	if _, ok := userdata.Dict_to_file_struct[filename]; ok == false {
		return "", errors.New(strings.ToTitle("The file never belonged to you"))
	}

	tmp := jacket.Receiver_dict[userdata.Dict_to_file_struct[filename]]
	if tmp == nil {
		new := make([]string,1)
		jacket.Receiver_dict[userdata.Dict_to_file_struct[filename]] = append(new, recipient)
	} else {
		jacket.Receiver_dict[userdata.Dict_to_file_struct[filename]] = append(tmp, recipient)
	}

	if _, ok := jacket.Direct_Children[recipient]; ok == false {
		jacket.Direct_Children[recipient] = true
	}
	if len(queried_jacket) < 80 {
		return "", errors.New((strings.ToTitle("Jacket struct broken")))
	}
	userdata.UpdateUser(jacket, queried_jacket[:80])

	recipient_pk_id, e := userlib.HMACEval(make([]byte, 16), []byte(recipient+"public"))
	if e != nil {
		return "", e
	}
	recipient_public_key, ok := userlib.KeystoreGet(string(recipient_pk_id))
	if ok == false {
		return "", errors.New(strings.ToTitle("Could not fetch recipient public key"))
	}

	unlock_key, e := userlib.HashKDF(userdata.EncLong, []byte("shellshare")) //one-time-use
	if e != nil {
		return "", e
	}
	unlock_key = unlock_key[:16]

	file_uuid := userdata.Dict_to_file_struct[filename]
	file_verify := userdata.Dict_to_file_verify[filename]
	file_decrypt := userdata.Dict_to_file_decrypt[filename]

	js_verify, e := json.Marshal(file_verify)
	if e != nil {
		return "", e
	}
	access_token := userlib.SymEnc(unlock_key, userlib.RandomBytes(16), append(append([]byte(file_uuid.String()), file_decrypt...), js_verify...))
	enc_key, e := userlib.PKEEnc(recipient_public_key, unlock_key)
	if e != nil {
		return "", e
	}
	access_token = append(enc_key, access_token...)
	signature, e := userlib.DSSign(userdata.Sign, append([]byte(userdata.Username), access_token...))
	if e != nil {
		return "", e
	}

	return string(signature)+string(access_token), nil
}

// Note recipient's filename can be different from the sender's filename.
// The recipient should not be able to discover the sender's view on
// what the filename even is!  However, the recipient must ensure that
// it is authentically from the sender.
func (userdata *User) ReceiveFile(filename string, sender string,
	magic_string string) error {
	// Check integrity and sync with the current userstruct
	jacket, queried_jacket, e := userdata.SyncUser()
	if e != nil {
		return e
	}
	if _, ok := userdata.Dict_to_file_struct[filename]; ok == true {
		return errors.New(strings.ToTitle("You already have a file under this name"))
	}


	if len([]byte(magic_string)) < 256 {
		return errors.New(strings.ToTitle("Magic string is broken"))
	}
	sus_signature := []byte(magic_string)[:256]
	enc := []byte(magic_string)[256:]

	// 1. Verify Integrity
	sender_verify_key, e := GetVerifyKey(sender)
	if e != nil {
    return e
  }
	if userlib.DSVerify(sender_verify_key, append([]byte(sender), enc...), sus_signature) != nil {
		return errors.New(strings.ToTitle("Access code/Sender integrity has been compromised"))
	}

	// 2. Decode access token for uuid of shell struct
	access_token := magic_string[256:]
	if len(access_token) < 256 {
		return errors.New(strings.ToTitle("Access token is broken"))
	}
	unlock_key, e := userlib.PKEDec(userdata.SK, []byte(access_token[:256]))
	if e != nil {

    return e
  }
	file_stuff := userlib.SymDec(unlock_key, []byte(access_token[256:]))
	if len(file_stuff) < 52 {
		return errors.New(strings.ToTitle("File sharing is broken"))
	}
	file_uuid, e := uuid.Parse(string(file_stuff[:36]))
	e = userdata.SyncFileForReceive(filename, string(file_stuff[:36]))
	if e != nil {
		return e
	}
	file_decrypt := file_stuff[36:52]
	var fv userlib.DSVerifyKey
	file_verify := &fv
	json.Unmarshal(file_stuff[52:], file_verify)
	if e != nil {
		return e
	}

	_, ok := userdata.Dict_to_file_struct[filename]
	if ok == false {
		userdata.Dict_to_file_struct[filename] = file_uuid
	}
	userdata.Dict_to_file_verify[filename] = *file_verify
	userdata.Dict_to_file_decrypt[filename] = file_decrypt

	jacket, queried_jacket, e = userdata.SyncUser()
	if e != nil {
		return e
	}

	// 3. Update user
	if len(queried_jacket) < 80 {
		return errors.New((strings.ToTitle("Jacket struct broken")))
	}
	userdata.UpdateUser(jacket, queried_jacket[:80])
	return nil
}


// Removes target user's access.
func (userdata *User) RevokeFile(filename string, target_username string) (err error) {
	// Check integrity and sync with the current userstruct
	//ti := time.Now()
	e := userdata.SyncFile()
	if e != nil {
		return e
	}
	jacket, queried_jacket, e := userdata.SyncUser()
	if e != nil {
		return e
	}

	// Check if it is a direct child
	if _, ok := jacket.Direct_Children[target_username]; ok == false {
		return errors.New(strings.ToTitle("Target is not a direct child"))
	}

	target_array := jacket.Receiver_dict[userdata.Dict_to_file_struct[filename]]
	for i, element := range target_array {
		if element == target_username {
			target_array[i] = target_array[len(target_array) - 1]
			jacket.Receiver_dict[userdata.Dict_to_file_struct[filename]] = target_array[:len(target_array)-1]
		}
	}
	if len(queried_jacket) < 80 {
		return errors.New((strings.ToTitle("Jacket struct broken")))
	}
	userdata.UpdateUser(jacket, queried_jacket[:80])

	// Move file head
	new_file_loc := uuid.New()
	curr_file_loc := userdata.Dict_to_file_struct[filename]
	curr_file_pkg, ok := userlib.DatastoreGet(curr_file_loc)
	if ok == false {
		return errors.New(strings.ToTitle("Cannot fetch from datastore"))
	}
	userlib.DatastoreSet(new_file_loc, curr_file_pkg)

	jacket, queried_jacket, e = userdata.SyncUser()
	if e != nil {
		return e
	}
	orig := userdata.Dict_to_original_file_struct[filename]
	e = UpdateShellFromRoot(userdata.Username, userdata.Dict_to_file_struct[filename], userdata.Sign, new_file_loc, orig, userdata.Username)
	if e != nil {
		return e
	}

	return nil
}

func UpdateShellFromRoot(node_name string, file_id uuid.UUID, creator_sk userlib.DSSignKey, new_address uuid.UUID, original_address uuid.UUID, creator string) error {
 	 // Get the shell struct of the root
	 user_uuid := GetUserUUID(node_name)
	 queried_jacket, ok := userlib.DatastoreGet(user_uuid)
	 if ok == false {
 		return errors.New(strings.ToTitle("Cannot fetch from datastore"))
 	 }
	 if len(queried_jacket) < 336 {
	 		return errors.New(strings.ToTitle("Jacket struct broken"))
	 }
	 sus_signature := queried_jacket[80:336]
 	 jacket_ser := queried_jacket[336:]
 	 var j Jacket
 	 jacket := &j
 	 json.Unmarshal(queried_jacket[336:], jacket)

 	 usr_verify_key, e := GetVerifyKey(node_name)
 	 if e != nil {
     return e
    }
 	 error_user := userlib.DSVerify(usr_verify_key, jacket_ser, sus_signature)
 	 if len(jacket.LastCreator) != 0 {
 	 creator_verify_key, e := GetVerifyKey(jacket.LastCreator)
 	 if e != nil {
 	   return e
 	 }
 	 error_creator := userlib.DSVerify(creator_verify_key, jacket_ser, sus_signature)
 	 if !(error_user == nil || error_creator == nil)  {
 		 return errors.New(strings.ToTitle("Jacket struct integrity has been compromised"))
 	 }
 	 } else {
 		 if error_user != nil {
 			return errors.New(strings.ToTitle("Jacket struct integrity has been compromised"))
 		}
 	 }
	 receiverdict := jacket.Receiver_dict
	 names := make([]string, 0, len(receiverdict[file_id]))

	 e = UpdateShell(node_name, file_id, user_uuid, jacket, queried_jacket, creator_sk, new_address, original_address, creator)
	 if e != nil {
		 return e
	 }

	 for _, v := range receiverdict[file_id] {
		 names = append(names, v)
	 }
	 for i := 0; i < len(names); i++ {
	 		UpdateShellFromRoot(names[i], file_id, creator_sk, new_address, original_address, creator)
	 }
	 return nil
}
func UpdateShell(username string, file_id uuid.UUID, user_uuid uuid.UUID, jacket *Jacket, queried_jacket []byte, creator_sk userlib.DSSignKey, new_address uuid.UUID, original_address uuid.UUID, creator string) error {
	// Get shell struct
	// Mark the updated field as True
	pk_id, e := userlib.HMACEval(make([]byte, 16), []byte(username+"public"))
	if e != nil {
		return e
	}
	public_key, ok := userlib.KeystoreGet(string(pk_id))
	if ok == false {
		return errors.New(strings.ToTitle("Could not fetch verify key for "+username))
	}
	enc_new, e := userlib.PKEEnc(public_key, append([]byte(original_address.String()), []byte(new_address.String())...))
	if e != nil {
		return e
	}


	if jacket.UpdatedAddress == nil {
		jacket.UpdatedAddress = make(map[uuid.UUID][]byte)
	}
	jacket.UpdatedAddress[file_id] = enc_new


	jacket.LastCreator = creator
	jacket_ser, e := json.Marshal(jacket)
	if e != nil {
		return e
	}
	signature, e := userlib.DSSign(creator_sk, jacket_ser)
	if e != nil {
		return e
	}

	if len(queried_jacket) < 80 {
		userlib.DebugMsg(strings.ToTitle("Jacket struct broken"))
	}
	prefix := queried_jacket[:80]
	to_upload := append(append(prefix, signature...), jacket_ser...)
	userlib.DatastoreSet(user_uuid, to_upload)

	return nil
}
