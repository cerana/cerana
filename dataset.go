package main

import "fmt"

const (
	DatasetFilesystem = "filesystem"
	DatasetSnapshot   = "snapshot"
	DatasetVolume     = "volume"
)

type Dataset struct {
	Name          string
	Origin        string
	Used          uint64
	Avail         uint64
	Mountpoint    string
	Compression   string
	Type          string
	Written       uint64
	Volsize       uint64
	Usedbydataset uint64
	Logicalused   uint64
	Quota         uint64
}

type ds struct {
	DMUObjsetStats *dmuObjsetStats `nv:"dmu_objset_stats"`
	Name           string          `nv:"name"`
	Properties     *dsProperties   `nv:"properties"`
}

type dmuObjsetStats struct {
	CreationTxg  uint64 `nv:"dds_creation_txg"`
	Guid         uint64 `nv:"dds_guid"`
	Inconsistent bool   `nv:"dds_inconsistent"`
	IsSnapshot   bool   `nv:"dds_is_snapshot"`
	NumClones    uint64 `nv:"dds_num_clonse"`
	Origin       string `nv:"dds_origin"`
	Type         string `nv:"dds_type"`
}

type dsProperties struct {
	Available            propUint64           `nv:"available"`
	Clones               clones               `nv:"clones"`
	Compression          propStringWithSource `nv:"compression"`
	CompressRatio        propUint64           `nv:"compressratio"`
	CreateTxg            propUint64           `nv:"createtxg"`
	Creation             propUint64           `nv:"creation"`
	DeferDestroy         propUint64           `nv:"defer_destroy"`
	Guid                 propUint64           `nv:"guid"`
	LogicalReferenced    propUint64           `nv:"logicalreferenced"`
	LogicalUsed          propUint64           `nv:"logicalused"`
	Mountpoint           propStringWithSource `nv:"mountpoint"`
	ObjsetID             propUint64           `nv:"objsetid"`
	Origin               propString           `nv:"origin"`
	Quota                propUint64WithSource `nv:"quota"`
	RefCompressRatio     propUint64           `nv:"refcompressratio"`
	RefQuota             propUint64WithSource `nv:"refquota"`
	RefReservation       propUint64WithSource `nv:"refreservation"`
	Referenced           propUint64           `nv:"referenced"`
	Reservation          propUint64WithSource `nv:"reservation"`
	Type                 propUint64           `nv:"type"`
	Unique               propUint64           `nv:"unique"`
	Used                 propUint64           `nv:"used"`
	UsedByChildren       propUint64           `nv:"usedbychildren"`
	UsedByDataset        propUint64           `nv:"usedbydataset"`
	UsedByRefReservation propUint64           `nv:"usedbyrefreservation"`
	UsedBySnapshots      propUint64           `nv:"usedbysnapshots"`
	UserAccounting       propUint64           `nv:"useraccounting"`
	UserRefs             propUint64           `nv:"userrefs"`
	Volsize              propUint64           `nv:"volsize"`
	VolBlockSize         propUint64           `nv:"volblocksize"`
	Written              propUint64           `nv:"written"`
}

type clones struct {
	Value map[string]bool `nv:"value"`
}

type propUint64 struct {
	Value uint64 `nv:"value"`
}

type propUint64WithSource struct {
	Source string `nv:"source"`
	Value  uint64 `nv:"value"`
}

type propString struct {
	Value string `nv:"value"`
}

type propStringWithSource struct {
	Source string `nv:"source"`
	Value  string `nv:"value"`
}

func dsToDataset(in *ds) *Dataset {
	var dsType string
	if in.DMUObjsetStats.IsSnapshot {
		dsType = DatasetSnapshot
	} else if dmuType(in.Properties.Type.Value) == dmuTypes["zvol"] {
		dsType = DatasetVolume
	} else {
		dsType = DatasetFilesystem
	}

	compression := in.Properties.Compression.Value
	if compression == "" {
		compression = "off"
	}

	mountpoint := in.Properties.Mountpoint.Value
	if mountpoint == "" && dsType != DatasetSnapshot {
		mountpoint = fmt.Sprintf("/%s", in.Name)
	}

	return &Dataset{
		Name:          in.Name,
		Origin:        in.Properties.Origin.Value,
		Used:          in.Properties.Used.Value,
		Avail:         in.Properties.Available.Value,
		Mountpoint:    mountpoint,
		Compression:   compression,
		Type:          dsType,
		Written:       in.Properties.Available.Value,
		Volsize:       in.Properties.Volsize.Value,
		Usedbydataset: in.Properties.UsedByDataset.Value,
		Logicalused:   in.Properties.LogicalUsed.Value,
		Quota:         in.Properties.Quota.Value,
	}
}

func Datasets(name string) ([]*Dataset, error) {
	types := map[string]bool{
		"all": true,
	}
	dss, err := list(name, types, false, 0)
	if err != nil {
		return nil, err
	}

	datasets := make([]*Dataset, len(dss))
	for i, ds := range dss {
		datasets[i] = dsToDataset(ds)
	}

	return datasets, nil
}
