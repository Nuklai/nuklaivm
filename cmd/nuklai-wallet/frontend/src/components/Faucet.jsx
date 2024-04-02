import {
  CheckCircleTwoTone,
  CloseCircleTwoTone,
  FundOutlined,
  LoadingOutlined,
  RedoOutlined
} from '@ant-design/icons'
import {
  Alert,
  App,
  Avatar,
  Button,
  Card,
  Divider,
  List,
  Space,
  Spin,
  Typography
} from 'antd'
import { useCallback, useEffect, useState } from 'react'
import {
  GetFaucetSolutions,
  StartFaucetSearch
} from '../../wailsjs/go/main/App'

const { Text, Title, Paragraph } = Typography
const antIcon = <LoadingOutlined style={{ fontSize: 24 }} spin />

const Faucet = () => {
  const { message } = App.useApp()
  const [loaded, setLoaded] = useState(false)
  const [search, setSearch] = useState(null)
  const [solutions, setSolutions] = useState([])

  const startSearch = async () => {
    try {
      const newSearch = await StartFaucetSearch()
      setSearch(newSearch)
    } catch (error) {
      message.error('Failed to start the search. Please try again.')
    }
  }

  const getFaucetSolutions = useCallback(async () => {
    try {
      const faucetSolutions = await GetFaucetSolutions()
      faucetSolutions.Alerts?.forEach((alert) => {
        message[alert.Type](alert.Content)
      })
      setSearch(faucetSolutions.CurrentSearch)
      setSolutions(faucetSolutions.PastSearches)
      setLoaded(true)
    } catch (error) {
      message.error(
        'Failed to fetch faucet solutions. Please refresh the page.'
      )
    }
  }, [message])

  useEffect(() => {
    getFaucetSolutions()
    const interval = setInterval(getFaucetSolutions, 10000) // Refresh every 10 seconds

    return () => clearInterval(interval)
  }, [getFaucetSolutions])

  return (
    <Space direction='vertical' size='large' style={{ width: '100%' }}>
      <Card bordered={false}>
        <Title level={3}>Token Faucet</Title>
        <Paragraph>
          This faucet provides test tokens for users on the Nuklai network. To
          request tokens, solve a simple computational puzzle to demonstrate
          your commitment and help prevent abuse.
        </Paragraph>
        {!loaded ? (
          <Spin indicator={antIcon} />
        ) : (
          <>
            {search ? (
              <Space direction='vertical' size='middle'>
                <Spin indicator={antIcon} />
                <Text>Your request is being processed...</Text>
                <Button type='default' icon={<RedoOutlined spin />} disabled>
                  Request Pending
                </Button>
              </Space>
            ) : (
              <Button
                type='primary'
                icon={<FundOutlined />}
                onClick={startSearch}
              >
                Request Tokens
              </Button>
            )}
          </>
        )}
      </Card>

      {solutions.length > 0 && (
        <>
          <Divider>Past Requests</Divider>
          <List
            itemLayout='horizontal'
            dataSource={solutions}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  avatar={
                    item.Err ? (
                      <Avatar
                        icon={<CloseCircleTwoTone twoToneColor='#f5222d' />}
                      />
                    ) : (
                      <Avatar
                        icon={<CheckCircleTwoTone twoToneColor='#52c41a' />}
                      />
                    )
                  }
                  title={
                    <Text>
                      {item.Err ? 'Request Failed' : 'Tokens Received'}
                    </Text>
                  }
                  description={
                    <>
                      <Text strong>Solution:</Text> {item.Solution} <br />
                      <Text strong>Salt:</Text> {item.Salt} <br />
                      <Text strong>Difficulty:</Text> {item.Difficulty} <br />
                      <Text strong>Attempts:</Text> {item.Attempts} <br />
                      {item.Err ? (
                        <Alert message={item.Err} type='error' showIcon />
                      ) : (
                        <>
                          <Text strong>Received:</Text> {item.Amount} <br />
                          <Text strong>Transaction ID:</Text> {item.TxID}
                        </>
                      )}
                    </>
                  }
                />
              </List.Item>
            )}
          />
        </>
      )}
    </Space>
  )
}

export default Faucet
