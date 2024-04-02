// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import { InfoCircleOutlined } from '@ant-design/icons'
import { Area, Line } from '@ant-design/plots'
import {
  App,
  Card,
  Col,
  Divider,
  List,
  Pagination,
  Popover,
  Row,
  Typography
} from 'antd'
import { useEffect, useState } from 'react'
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
  const blocksPerPage = 10 // Adjust based on your preference
  const [totalBlocks, setTotalBlocks] = useState(0) // This state will hold the total number of blocks

  useEffect(() => {
    const fetchData = async () => {
      try {
        const chainId = await GetChainID()
        setChainID(chainId)

        const latestBlocks = await GetLatestBlocks(currentPage, blocksPerPage)
        setBlocks(latestBlocks)

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
    }

    fetchData()

    // Set up the interval to refetch data every 10 seconds
    const intervalId = setInterval(() => {
      fetchData()
    }, 10000) // 10000 milliseconds = 10 seconds

    // Clear the interval when the component is unmounted
    return () => clearInterval(intervalId)
  }, [currentPage, message]) // The dependencies array ensures fetchData is called when currentPage changes or message is used

  const handlePageChange = (page) => {
    setCurrentPage(page)
  }

  return (
    <>
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
              <Text strong>Active Accounts:</Text> # of accounts issusing
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
      <Row gutter={16}>
        <Col span={8}>
          <Card title='Transactions Per Second' bordered={true}>
            <Area
              data={transactionStats}
              xField={'Timestamp'}
              yField={'Count'}
              autoFit={true}
              height={200}
              animation={false}
              xAxis={{ tickCount: 0 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title='Active Accounts' bordered={true}>
            <Area
              data={accountStats}
              xField={'Timestamp'}
              yField={'Count'}
              autoFit={true}
              height={200}
              animation={false}
              xAxis={{ tickCount: 0 }}
            />
          </Card>
        </Col>
        <Col span={8}>
          <Card title='Unit Prices' bordered={true}>
            <Line
              data={unitPrices}
              xField={'Timestamp'}
              yField={'Count'}
              seriesField={'Category'}
              autoFit={true}
              height={200}
              animation={false}
              legend={false}
              xAxis={{ tickCount: 0 }}
            />
          </Card>
        </Col>
      </Row>
      <Divider orientation='center'>
        <Popover
          content={
            <div>
              <Text italic>
                Recent activity for NuklaiNet (ChainID: {chainID})
              </Text>
              <br />
              <br />
              <Text strong>Timestamp:</Text> Time that block was created
              <br />
              <Text strong>Transactions:</Text> # of successful transactions in
              block
              <br />
              <Text strong>Units Consumed:</Text> # of HyperSDK fee units
              consumed
              <br />
              <Text strong>State Root:</Text> Merkle root of State at start of
              block execution
              <br />
              <Text strong>Block Size:</Text> Size of block in bytes
              <br />
              <Text strong>Accept Latency:</Text> Difference between block
              creation and block acceptance
            </div>
          }
        >
          Blocks <InfoCircleOutlined />
        </Popover>
      </Divider>

      <List
        bordered
        dataSource={blocks}
        renderItem={(item) => (
          <List.Item key={item.ID}>
            <div>
              <Title level={3} style={{ display: 'inline' }}>
                {item.Height}
              </Title>{' '}
              <Text type='secondary'>{item.ID}</Text>
            </div>
            <Text strong>Timestamp:</Text> {item.Timestamp}
            <br />
            <Text strong>Transactions:</Text> {item.Txs}
            {item.Txs > 0 && (
              <Text italic type='danger'>
                {' '}
                (failed: {item.FailTxs})
              </Text>
            )}
            <br />
            <Text strong>Units Consumed:</Text> {item.Consumed}
            <br />
            <Text strong>State Root:</Text> {item.StateRoot}
            <br />
            <Text strong>Block Size:</Text> {item.Size}
            <br />
            <Text strong>Accept Latency:</Text> {item.Latency}ms
          </List.Item>
        )}
      />
      <Pagination
        current={currentPage}
        onChange={handlePageChange}
        total={totalBlocks} // Use the state holding the total number of blocks for pagination
        pageSize={blocksPerPage}
        showSizeChanger={false}
      />
    </>
  )
}

export default Explorer
