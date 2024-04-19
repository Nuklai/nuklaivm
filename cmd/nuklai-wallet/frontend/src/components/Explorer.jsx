// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import { InfoCircleOutlined } from '@ant-design/icons'
import { Area, Line } from '@ant-design/plots'
import {
  App,
  Card,
  Col,
  Descriptions,
  Divider,
  List,
  Pagination,
  Popover,
  Row,
  Tag,
  Typography
} from 'antd'
import { useCallback, useEffect, useState } from 'react'
import {
  GetAccountStats,
  GetChainID,
  GetLatestBlocks,
  GetTotalBlocks,
  GetTransactionStats,
  GetUnitPrices
} from '../../wailsjs/go/main/App'

const { Title, Text } = Typography

const Explorer = () => {
  const { message } = App.useApp()
  const [blocks, setBlocks] = useState([])
  const [transactionStats, setTransactionStats] = useState([])
  const [accountStats, setAccountStats] = useState([])
  const [unitPrices, setUnitPrices] = useState([])
  const [chainID, setChainID] = useState('')
  const [currentPage, setCurrentPage] = useState(1)
  const blocksPerPage = 10
  const [totalBlocks, setTotalBlocks] = useState(0)

  // Using useCallback to memoize fetchData function to avoid infinite loop in useEffect
  const fetchData = useCallback(async () => {
    try {
      const chainId = await GetChainID()
      setChainID(chainId)

      const latestBlocks = await GetLatestBlocks(currentPage, blocksPerPage)
      if (latestBlocks && latestBlocks.length > 0) {
        setBlocks(latestBlocks)
      }

      const total = await GetTotalBlocks()
      setTotalBlocks(total)

      const [transStats, accStats, prices] = await Promise.all([
        GetTransactionStats(),
        GetAccountStats(),
        GetUnitPrices()
      ])

      setTransactionStats(transStats)
      setAccountStats(accStats)
      setUnitPrices(prices)
    } catch (error) {
      message.error(error.toString())
    }
  }, [currentPage, message]) // Added currentPage as a dependency

  useEffect(() => {
    fetchData()

    // Set up the interval to refetch data every 5 seconds
    const intervalId = setInterval(fetchData, 5000)

    // Clear the interval when the component is unmounted
    return () => clearInterval(intervalId)
  }, [fetchData]) // Now depends on fetchData which includes currentPage

  const handlePageChange = (page) => {
    setCurrentPage(page)
  }

  return (
    <>
      {/* Metrics and stats section */}
      <Divider orientation='center'>
        <Popover
          content={
            <div>
              <Text italic>
                Collection of NuklaiNet Telemetry Over the Last 2 Minutes
              </Text>
              <br />
              <br />
              <Text strong>Transactions Per Second:</Text> # of transactions
              accepted per second
              <br />
              <Text strong>Active Accounts:</Text> # of accounts issuing
              transactions
              <br />
              <Text strong>Unit Prices:</Text> Price of each HyperSDK fee
              dimension (Bandwidth, Compute, Storage[Read], Storage[Allocate],
              Storage[Write])
            </div>
          }
        >
          Metrics <InfoCircleOutlined />
        </Popover>
      </Divider>

      {/* Charts section */}
      <Row gutter={16}>
        <Col span={8}>
          <Card title='Transactions Per Second' bordered>
            <Area
              data={transactionStats}
              xField='Timestamp'
              yField='Count'
              autoFit
              height={200}
              animation={false}
              xAxis={{ tickCount: 0 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title='Active Accounts' bordered>
            <Area
              data={accountStats}
              xField='Timestamp'
              yField='Count'
              autoFit
              height={200}
              animation={false}
              xAxis={{ tickCount: 0 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title='Unit Prices' bordered>
            <Line
              data={unitPrices}
              xField='Timestamp'
              yField='Count'
              seriesField='Category'
              autoFit
              height={200}
              animation={false}
              legend={false}
              xAxis={{ tickCount: 0 }}
            />
          </Card>
        </Col>
      </Row>

      {/* Blocks section */}
      <Divider orientation='center'>
        Recent activity for NuklaiNet (ChainID: {chainID})
      </Divider>

      <List
        bordered
        dataSource={blocks}
        renderItem={(item) => (
          <List.Item>
            <Card>
              <Descriptions title={`Block #${item.Height}`}>
                <Descriptions.Item label='ID'>{item.ID}</Descriptions.Item>
                <Descriptions.Item label='Timestamp'>
                  {item.Timestamp}
                </Descriptions.Item>
                <Descriptions.Item label='Transactions'>
                  {item.Txs}{' '}
                  <Tag color={item.FailTxs > 0 ? 'volcano' : 'green'}>
                    {item.FailTxs > 0
                      ? `Failed: ${item.FailTxs}`
                      : 'All Successful'}
                  </Tag>
                </Descriptions.Item>
                <Descriptions.Item label='Units Consumed'>
                  {item.Consumed}
                </Descriptions.Item>
                <Descriptions.Item label='State Root'>
                  {item.StateRoot}
                </Descriptions.Item>
                <Descriptions.Item label='Block Size'>
                  {item.Size}
                </Descriptions.Item>
                <Descriptions.Item label='Accept Latency'>
                  {item.Latency}ms
                </Descriptions.Item>
              </Descriptions>
            </Card>
          </List.Item>
        )}
      />

      {/* Pagination */}
      <Pagination
        current={currentPage}
        onChange={handlePageChange}
        total={totalBlocks}
        pageSize={blocksPerPage}
        showSizeChanger={false}
        style={{ marginTop: '20px', textAlign: 'center' }}
      />
    </>
  )
}

export default Explorer
