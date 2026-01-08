package models

import "fmt"

type RegionInfo struct {
	// region id
	RegionId string
	// 表示程序运行的网络是否为VPC网络
	Vpc bool
	// 终端地址信息
	Endpoint string
	// CA 证书文件路径
	CaFilePath string
}

// ToString 将RegionInfo转换为字符串表示
func (r *RegionInfo) ToString() string {
	return fmt.Sprintf("RegionInfo{RegionId: %s, Vpc: %t, Endpoint: %s, CaFilePath: %s}",
		r.RegionId, r.Vpc, r.Endpoint, r.CaFilePath)
}

// String 实现Stringer接口
func (r *RegionInfo) String() string {
	return r.ToString()
}

type RegionInfoExtend struct {
	*RegionInfo
	Elapsed   float64
	Reachable bool
}

// ToString 将RegionInfoExtend转换为字符串表示
func (r *RegionInfoExtend) ToString() string {
	return fmt.Sprintf("RegionInfoExtend{RegionInfo: %s, Elapsed: %f, Reachable: %t}",
		r.RegionInfo.ToString(), r.Elapsed, r.Reachable)
}

// String 实现Stringer接口
func (r *RegionInfoExtend) String() string {
	return r.ToString()
}

func NewRegionInfoWithRegionId(regionId string) *RegionInfo {
	return &RegionInfo{
		RegionId: regionId,
	}
}

func NewRegionInfoWithEndpoint(regionId string, endpoint string) *RegionInfo {
	return &RegionInfo{
		RegionId: regionId,
		Endpoint: endpoint,
	}
}

func NewRegionInfoWithVpcEndpoint(regionId string, vpc bool, endpoint string) *RegionInfo {
	return &RegionInfo{
		RegionId: regionId,
		Vpc:      vpc,
		Endpoint: endpoint,
	}
}

func NewRegionInfoWithCaFilePath(regionId string, endpoint string, caFilePath string) *RegionInfo {
	return &RegionInfo{
		RegionId:   regionId,
		Endpoint:   endpoint,
		CaFilePath: caFilePath,
	}
}
