package lib

import (
	"github.com/ks3sdklib/aws-sdk-go/aws"
	"github.com/ks3sdklib/aws-sdk-go/service/s3"
	. "gopkg.in/check.v1"
	"strings"
)

func (s *Ks3utilCommandSuite) TestQueryKs3Data(c *C) {
	c.Skip("Skip TestQueryKs3Data")
	// 获取指定桶在20250901-20250902期间的用量详情
	resp, err := client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
	c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)

	// 错误参数，BucketNames超过5个
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{"bucket1", "bucket2", "bucket3", "bucket4", "bucket5", "bucket6"},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
	c.Assert(strings.Contains(err.Error(), "The number of specified buckets can not exceed 5"), Equals, true)

	// 错误参数，跨月查询
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250831"),
		EndTime:     aws.String("20250901"),
		BucketNames: []string{bucket},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
	c.Assert(strings.Contains(err.Error(), "StartTime and EndTime should be in the same month"), Equals, true)

	// 错误参数，StartTime大于EndTime
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250902"),
		EndTime:     aws.String("20250901"),
		BucketNames: []string{bucket},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
	c.Assert(strings.Contains(err.Error(), "The EndTime should be later than the StartTime"), Equals, true)

	// 错误参数，StartTime和EndTime格式错误
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("202509010000"),
		EndTime:     aws.String("202509022359"),
		BucketNames: []string{bucket},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
	c.Assert(strings.Contains(err.Error(), "Invalid length of StartTime or EndTime"), Equals, true)

	// 获取指定桶在20250901-20250902期间指定的用量详情
	Ks3ProductList := []string{"DataSize", "NetworkFlowUp", "NetworkFlow", "CDNFlow", "ReplicationFlow", "RequestsGet", "RequestsPut", "RestoreSize", "TagNum", "BandwidthUp", "BandwidthDown", "NetBandwidthUp", "NetBandwidthDown", "CDNBandwidthDown", "IntranetBandwidthUp", "IntranetBandwidthDown", "IntranetFlowUp", "IntranetFlowDown", "ObjectNum"}
	for _, val := range Ks3ProductList {
		resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
			StartTime:   aws.String("20250901"),
			EndTime:     aws.String("20250902"),
			BucketNames: []string{bucket},
			Ks3Products: []string{val},
		})
		c.Assert(err, IsNil)
		c.Assert(*resp.StatusCode, Equals, int64(200))
		c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)
	}

	// 同时指定多个用量详情
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
		Ks3Products: Ks3ProductList,
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
	c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)

	// 指定一个错误的用量详情
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
		Ks3Products: []string{"Test"},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
	c.Assert(strings.Contains(err.Error(), "Invalid Ks3Product"), Equals, true)

	// 获取指定桶在20250901-20250902期间指定Object、Referer、IP、UA产生的流量
	TransferList := []string{"Object", "Referer", "IP", "UA"}
	for _, val := range TransferList {
		resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
			StartTime:   aws.String("20250901"),
			EndTime:     aws.String("20250902"),
			BucketNames: []string{bucket},
			Transfers:   []string{val},
		})
		c.Assert(err, IsNil)
		c.Assert(*resp.StatusCode, Equals, int64(200))
		c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)
	}

	// 指定一个错误的流量维度
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
		Transfers:   []string{"Test"},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))

	// 同时指定多个流量维度
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
		Transfers:   TransferList,
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
	c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)

	// 获取指定桶在20250901-20250902期间指定Object、Referer、IP、UA产生的请求次数
	RequestList := []string{"Object", "Referer", "IP", "UA"}
	for _, val := range RequestList {
		resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
			StartTime:   aws.String("20250901"),
			EndTime:     aws.String("20250902"),
			BucketNames: []string{bucket},
			Requests:    []string{val},
		})
		c.Assert(err, IsNil)
		c.Assert(*resp.StatusCode, Equals, int64(200))
		c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)
	}

	// 同时指定多个请求次数维度
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
		Requests:    RequestList,
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
	c.Assert(len(resp.Ks3DataResult.Data.Buckets), Equals, 2)

	// 指定一个错误的请求次数维度
	resp, err = client.QueryKs3Data(&s3.QueryKs3DataInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		BucketNames: []string{bucket},
		Requests:    []string{"Test"},
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
}

func (s *Ks3utilCommandSuite) TestQueryBucketRank(c *C) {
	c.Skip("Skip TestQueryBucketRank")
	// 获取20250901-20250902期间，按数据量排序的前200个桶
	resp, err := client.QueryBucketRank(&s3.QueryBucketRankInput{
		StartTime: aws.String("20250901"),
		EndTime:   aws.String("20250902"),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))

	// 获取20250901-20250902期间，按数据量排序的前500个桶
	resp, err = client.QueryBucketRank(&s3.QueryBucketRankInput{
		StartTime: aws.String("20250901"),
		EndTime:   aws.String("20250902"),
		Number:    aws.Long(500),
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))

	// 设置Number超过500
	resp, err = client.QueryBucketRank(&s3.QueryBucketRankInput{
		StartTime: aws.String("20250901"),
		EndTime:   aws.String("20250902"),
		Number:    aws.Long(501),
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))
	c.Assert(strings.Contains(err.Error(), "Param Number should be [1, 500]"), Equals, true)

	// 错误参数，StartTime大于EndTime
	resp, err = client.QueryBucketRank(&s3.QueryBucketRankInput{
		StartTime: aws.String("20250902"),
		EndTime:   aws.String("20250901"),
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))

	// 错误参数，StartTime和EndTime格式错误
	resp, err = client.QueryBucketRank(&s3.QueryBucketRankInput{
		StartTime: aws.String("202509010000"),
		EndTime:   aws.String("202509022359"),
	})
	c.Assert(err, NotNil)
	c.Assert(*resp.StatusCode, Equals, int64(400))

	// 获取指定桶在20250901-20250902期间，按指定用量排序的前200个桶
	Ks3ProductList := []string{"DataSize", "Flow", "RequestsGet", "RequestsPut"}
	for _, val := range Ks3ProductList {
		resp, err = client.QueryBucketRank(&s3.QueryBucketRankInput{
			StartTime:   aws.String("20250901"),
			EndTime:     aws.String("20250902"),
			Ks3Products: []string{val},
		})
		c.Assert(err, IsNil)
		c.Assert(*resp.StatusCode, Equals, int64(200))
	}

	// 同时指定多个用量详情
	resp, err = client.QueryBucketRank(&s3.QueryBucketRankInput{
		StartTime:   aws.String("20250901"),
		EndTime:     aws.String("20250902"),
		Ks3Products: Ks3ProductList,
	})
	c.Assert(err, IsNil)
	c.Assert(*resp.StatusCode, Equals, int64(200))
}
