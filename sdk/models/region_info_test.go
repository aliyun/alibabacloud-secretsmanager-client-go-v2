package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

// TestRegionInfo_ToString 测试RegionInfo的ToString方法
func TestRegionInfo_ToString(t *testing.T) {
	// 测试基本RegionInfo
	regionInfo := &RegionInfo{
		RegionId:   "cn-hangzhou",
		Vpc:        false,
		Endpoint:   "kms.cn-hangzhou.aliyuncs.com",
		CaFilePath: "/path/to/ca.pem",
	}

	expected := "RegionInfo{RegionId: cn-hangzhou, Vpc: false, Endpoint: kms.cn-hangzhou.aliyuncs.com, CaFilePath: /path/to/ca.pem}"
	actual := regionInfo.ToString()
	assert.Equal(t, expected, actual, "ToString should return correct string representation")

	// 测试VPC类型的RegionInfo
	vpcRegionInfo := &RegionInfo{
		RegionId:   "cn-shanghai",
		Vpc:        true,
		Endpoint:   "kms-vpc.cn-shanghai.aliyuncs.com",
		CaFilePath: "",
	}

	expected = "RegionInfo{RegionId: cn-shanghai, Vpc: true, Endpoint: kms-vpc.cn-shanghai.aliyuncs.com, CaFilePath: }"
	actual = vpcRegionInfo.ToString()
	assert.Equal(t, expected, actual, "ToString should return correct string representation for VPC region")

	t.Log("RegionInfo ToString test passed")
}

// TestRegionInfoExtend_ToString 测试RegionInfoExtend的ToString方法
func TestRegionInfoExtend_ToString(t *testing.T) {
	// 创建基础RegionInfo
	regionInfo := &RegionInfo{
		RegionId:   "cn-hangzhou",
		Vpc:        false,
		Endpoint:   "kms.cn-hangzhou.aliyuncs.com",
		CaFilePath: "/path/to/ca.pem",
	}

	// 创建RegionInfoExtend
	regionInfoExtend := &RegionInfoExtend{
		RegionInfo: regionInfo,
		Elapsed:    100.5,
		Reachable:  true,
	}

	expected := "RegionInfoExtend{RegionInfo: RegionInfo{RegionId: cn-hangzhou, Vpc: false, Endpoint: kms.cn-hangzhou.aliyuncs.com, CaFilePath: /path/to/ca.pem}, Escaped: 100.500000, Reachable: true}"
	actual := regionInfoExtend.ToString()
	assert.Equal(t, expected, actual, "ToString should return correct string representation for RegionInfoExtend")

	t.Log("RegionInfoExtend ToString test passed")
}

// TestNewRegionInfoConstructors 测试所有RegionInfo构造函数
func TestNewRegionInfoConstructors(t *testing.T) {
	// 测试NewRegionInfoWithRegionId
	region1 := NewRegionInfoWithRegionId("cn-hangzhou")
	assert.Equal(t, "cn-hangzhou", region1.RegionId, "Region ID should match")
	assert.False(t, region1.Vpc, "VPC should be false by default")
	assert.Empty(t, region1.Endpoint, "Endpoint should be empty by default")
	assert.Empty(t, region1.CaFilePath, "CaFilePath should be empty by default")

	// 测试NewRegionInfoWithEndpoint
	region2 := NewRegionInfoWithEndpoint("cn-shanghai", "kms.cn-shanghai.aliyuncs.com")
	assert.Equal(t, "cn-shanghai", region2.RegionId, "Region ID should match")
	assert.False(t, region2.Vpc, "VPC should be false by default")
	assert.Equal(t, "kms.cn-shanghai.aliyuncs.com", region2.Endpoint, "Endpoint should match")
	assert.Empty(t, region2.CaFilePath, "CaFilePath should be empty by default")

	// 测试NewRegionInfoWithVpcEndpoint
	region3 := NewRegionInfoWithVpcEndpoint("cn-beijing", true, "kms-vpc.cn-beijing.aliyuncs.com")
	assert.Equal(t, "cn-beijing", region3.RegionId, "Region ID should match")
	assert.True(t, region3.Vpc, "VPC should be true")
	assert.Equal(t, "kms-vpc.cn-beijing.aliyuncs.com", region3.Endpoint, "Endpoint should match")
	assert.Empty(t, region3.CaFilePath, "CaFilePath should be empty by default")

	// 测试NewRegionInfoWithCaFilePath
	region4 := NewRegionInfoWithCaFilePath("cn-shenzhen", "kms.cn-shenzhen.aliyuncs.com", "/path/to/ca.pem")
	assert.Equal(t, "cn-shenzhen", region4.RegionId, "Region ID should match")
	assert.False(t, region4.Vpc, "VPC should be false by default")
	assert.Equal(t, "kms.cn-shenzhen.aliyuncs.com", region4.Endpoint, "Endpoint should match")
	assert.Equal(t, "/path/to/ca.pem", region4.CaFilePath, "CaFilePath should match")

	t.Log("NewRegionInfo constructors test passed")
}
