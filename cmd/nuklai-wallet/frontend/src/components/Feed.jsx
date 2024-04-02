import { InfoCircleOutlined } from '@ant-design/icons'
import {
  App,
  Button,
  Divider,
  Drawer,
  Form,
  Input,
  InputNumber,
  List,
  Popover,
  Select,
  Typography
} from 'antd'
import React, { useEffect, useState } from 'react'
import {
  GetBalance,
  GetFeed,
  GetFeedInfo,
  Message,
  OpenLink,
  Transfer as Send
} from '../../wailsjs/go/main/App'
import FundsCheck from './FundsCheck'

const { Title, Text } = Typography

const Feed = () => {
  const { message } = App.useApp()
  const [feed, setFeed] = useState([])
  const [feedInfo, setFeedInfo] = useState({})
  const [openCreate, setOpenCreate] = useState(false)
  const [createForm] = Form.useForm()
  const [openTip, setOpenTip] = useState(false)
  const [tipFocus, setTipFocus] = useState({})
  const [tipForm] = Form.useForm()
  const [balance, setBalance] = useState([])

  // Helper function to convert timestamp
  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  // Fetch data for the feed and user's balance
  useEffect(() => {
    const fetchData = async () => {
      const [feedData, feedInfoData, balances] = await Promise.all([
        GetFeed(),
        GetFeedInfo(),
        GetBalance()
      ])
      setFeed(feedData)
      setFeedInfo(feedInfoData)
      setBalance(
        balances.map((bal) => ({
          value: bal.ID,
          label: `${bal.Bal} ${bal.Symbol}`
        }))
      )
    }

    fetchData()
    const interval = setInterval(fetchData, 30000) // Refresh every 30 seconds
    return () => clearInterval(interval)
  }, [])

  // Handle creating a new post
  const onFinishCreate = async (values) => {
    setOpenCreate(false)
    message.loading({ content: 'Processing Transaction...', key: 'updatable' })
    try {
      await Message(values.Message, values.URL)
      message.success({
        content: 'Transaction Successful!',
        key: 'updatable',
        duration: 2
      })
      // Refetch the feed after posting
      setFeed(await GetFeed())
    } catch (error) {
      message.error({
        content: error.toString(),
        key: 'updatable',
        duration: 2
      })
    }
  }

  // Handle sending a tip
  const onFinishTip = async (values) => {
    setOpenTip(false)
    message.loading({ content: 'Processing Transaction...', key: 'updatable' })
    try {
      await Send(
        values.Asset,
        tipFocus.Address,
        values.Amount,
        `[${tipFocus.ID}]: ${values.Memo}`
      )
      message.success({
        content: 'Tip Sent Successfully!',
        key: 'updatable',
        duration: 2
      })
      // Update the balance after tipping
      setBalance(await GetBalance())
    } catch (error) {
      message.error({
        content: error.toString(),
        key: 'updatable',
        duration: 2
      })
    }
  }

  return (
    <>
      <div style={{ width: '60%', margin: 'auto' }}>
        <FundsCheck />
        <Divider orientation='center'>
          Posts
          <Popover
            content={
              <div>
                <p>
                  Because the fees are low on NuklaiNet, it is great for
                  micropayments.
                </p>
                <p>
                  This example allows anyone to pay the feed operator to post
                  content for everyone else to see.
                </p>
                <p>
                  If the amount of posts goes above the target/5 minutes, the
                  fee to post will increase.
                </p>
                <p>You can tip posters with any token you own!</p>
              </div>
            }
          >
            <InfoCircleOutlined />
          </Popover>
        </Divider>
        <Button
          type='primary'
          onClick={() => setOpenCreate(true)}
          disabled={!window.HasBalance}
        >
          Create Post
        </Button>
        <List
          itemLayout='vertical'
          size='large'
          dataSource={feed}
          renderItem={(item) => (
            <List.Item
              key={item.ID}
              actions={[
                <Button
                  onClick={() => {
                    setTipFocus(item)
                    setOpenTip(true)
                  }}
                >
                  Tip
                </Button>
              ]}
              extra={
                item.URLMeta?.Image && (
                  <img width={272} alt='thumbnail' src={item.URLMeta.Image} />
                )
              }
            >
              <List.Item.Meta
                title={
                  item.URLMeta ? (
                    <a onClick={() => OpenLink(item.URL)}>
                      {item.URLMeta.Title}
                    </a>
                  ) : (
                    <Text>{item.Message}</Text>
                  )
                }
                description={item.URLMeta?.Description}
              />
              <div>
                <Text strong>URL:</Text> {item.URL}
                <br />
                <Text strong>Message:</Text> {item.Message}
                <br />
                <Text strong>TxID:</Text> {item.ID}
                <br />
                <Text strong>Timestamp:</Text> {formatTimestamp(item.Timestamp)}
                <br />
                <Text strong>Fee:</Text> {item.Fee}
                <br />
                <Text strong>Actor:</Text> {item.Address}
              </div>
            </List.Item>
          )}
        />
      </div>

      <Drawer
        title='Create Post'
        placement='right'
        onClose={() => setOpenCreate(false)}
        open={openCreate}
      >
        <Form form={createForm} onFinish={onFinishCreate}>
          <Form.Item name='Message' rules={[{ required: true }]}>
            <Input placeholder='Enter your message' />
          </Form.Item>
          <Form.Item name='URL'>
            <Input placeholder='Add a link (optional)' />
          </Form.Item>
          <Button type='primary' htmlType='submit'>
            Post
          </Button>
        </Form>
      </Drawer>

      <Drawer
        title='Send Tip'
        placement='right'
        onClose={() => setOpenTip(false)}
        open={openTip}
      >
        <Form form={tipForm} onFinish={onFinishTip}>
          <Form.Item name='Asset' rules={[{ required: true }]}>
            <Select options={balance} placeholder='Select token' />
          </Form.Item>
          <Form.Item name='Amount' rules={[{ required: true }]}>
            <InputNumber placeholder='Amount' style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name='Memo'>
            <Input placeholder='Add a message (optional)' />
          </Form.Item>
          <Button type='primary' htmlType='submit'>
            Tip
          </Button>
        </Form>
      </Drawer>
    </>
  )
}

export default Feed
