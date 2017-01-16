package v1alpha1

type VirtStorageVol struct {
	Name string          `json:"name"`
	Pool VirtStoragePool `json:"pool"`
}

type VirtStoragePool struct {
	Dir   *VirtStoragePoolDir   `json:"dir"`
	NetFS *VirtStoragePoolNetFS `json:"netfs"`
	RBD   *VirtStoragePoolRBD   `json:"rbd"`
}

type VirtStoragePoolDir struct {
	Path string `json:"path"`
}

type VirtStoragePoolNetFS struct {
	Server string `json:"server"`
	Path   string `json:"path"`
}

type VirtStoragePoolRBD struct {
	Servers []string `json:"server"`
	Path    string   `json:"path"`
}
