import { PlusOutlined } from '@ant-design/icons'
import {
  App,
  Button,
  Card,
  Divider,
  Form,
  Input,
  InputNumber,
  Select,
  Space,
  Tooltip,
  Typography
} from 'antd'
import React, { useCallback, useEffect, useState } from 'react'
import {
  AddAddressBook,
  GetAddressBook,
  GetBalance,
  Transfer as Send
} from '../../wailsjs/go/main/App'
import FundsCheck from './FundsCheck'

const { Text } = Typography

const Transfer = () => {
  const { message } = App.useApp()
  const [balance, setBalance] = useState([])
  const [transferForm] = Form.useForm()
  const [addresses, setAddresses] = useState([])
  const [newNickname, setNewNickname] = useState('')
  const [newAddress, setNewAddress] = useState('')

  const getBalance = useCallback(async () => {
    const bals = await GetBalance()
    const parsedBalances = bals.map((bal) => ({
      value: bal.ID,
      label: bal.Bal
    }))
    setBalance(parsedBalances)
  }, [])

  const getAddresses = useCallback(async () => {
    const caddresses = await GetAddressBook()
    setAddresses(caddresses)
  }, [])

  const addAddress = async () => {
    try {
      await AddAddressBook(newNickname, newAddress)
      setNewNickname('')
      setNewAddress('')
      await getAddresses()
    } catch (e) {
      message.error(e.toString())
    }
  }

  const onFinishTransfer = async (values) => {
    const loadingMessageKey = 'processingTransaction' // Unique key for the loading message

    // Show the loading message without a duration, so it remains until manually updated or closed
    message.loading({
      content: 'Processing Transaction...',
      key: loadingMessageKey
    })

    try {
      // Convert amount to string to avoid parsing issues
      const amountAsString = values.Amount.toString()

      const start = new Date().getTime()
      // Use the string version of the amount for the transaction
      await Send(values.Asset, values.Address, amountAsString, values.Memo)
      const finish = new Date().getTime()

      // Calculate the duration in seconds
      const durationInSeconds = ((finish - start) / 1000).toFixed(2)

      // Update the message to success after the transaction is finalized and include the duration in seconds
      message.success({
        content: `Transaction Finalized (${durationInSeconds} seconds)`,
        key: loadingMessageKey, // Use the same key to update the existing message
        duration: 5 // Set how long the success message will stay (optional)
      })

      getBalance()

      // Reset the form fields after successful submission
      transferForm.resetFields()
    } catch (e) {
      // Update the message to error if the transaction fails
      message.error({
        content: e.toString(),
        key: loadingMessageKey,
        duration: 5
      }) // Optionally, set duration for the error message
    }
  }

  useEffect(() => {
    getBalance()
    getAddresses()
  }, [getBalance, getAddresses])

  return (
    <div style={{ maxWidth: '700px', margin: 'auto', padding: '20px' }}>
      <FundsCheck />
      <Card bordered>
        <Typography.Title level={4}>Send a Token</Typography.Title>
        <Form form={transferForm} onFinish={onFinishTransfer} layout='vertical'>
          <Form.Item
            name='Address'
            label='Recipient'
            rules={[{ required: true }]}
          >
            <Select
              showSearch
              optionFilterProp='children'
              placeholder='Select recipient'
              filterOption={(input, option) =>
                option.children.toLowerCase().indexOf(input.toLowerCase()) >= 0
              }
            >
              {addresses.map((addr) => (
                <Select.Option key={addr.Address} value={addr.Address}>
                  {addr.Name}: {addr.Address}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name='Asset' label='Token' rules={[{ required: true }]}>
            <Select placeholder='Select token'>
              {balance.map((bal) => (
                <Select.Option key={bal.value} value={bal.value}>
                  {bal.label}
                </Select.Option>
              ))}
            </Select>
          </Form.Item>
          <Form.Item name='Amount' label='Amount' rules={[{ required: true }]}>
            <InputNumber min={0} style={{ width: '100%' }} />
          </Form.Item>
          <Form.Item name='Memo' label='Memo (optional)'>
            <Input maxLength={256} />
          </Form.Item>
          <Form.Item>
            <Button
              type='primary'
              htmlType='submit'
              disabled={!window.HasBalance}
            >
              Send
            </Button>
          </Form.Item>
        </Form>
        <Divider />
        <Typography.Paragraph>
          <Text strong>Add a new address to the book:</Text>
        </Typography.Paragraph>
        <Space>
          <Input
            value={newNickname}
            onChange={(e) => setNewNickname(e.target.value)}
            placeholder='Nickname'
            style={{ width: 130 }}
          />
          <Input
            value={newAddress}
            onChange={(e) => setNewAddress(e.target.value)}
            placeholder='Address'
            style={{ width: 200 }}
          />
          <Tooltip title='Add to address book'>
            <Button
              icon={<PlusOutlined />}
              onClick={addAddress}
              disabled={!newNickname || !newAddress}
            />
          </Tooltip>
        </Space>
      </Card>
    </div>
  )
}

export default Transfer
