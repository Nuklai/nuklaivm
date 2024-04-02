import {
  App,
  Button,
  Card,
  Form,
  Input,
  InputNumber,
  Modal,
  Select,
  Space,
  Typography
} from 'antd'
import React, { useEffect, useState } from 'react'
import {
  CreateAsset,
  GetAddressBook,
  GetMyAssets,
  MintAsset
} from '../../wailsjs/go/main/App'
import FundsCheck from './FundsCheck'

const { Title } = Typography

const Mint = () => {
  const { message } = App.useApp()
  const [assets, setAssets] = useState([])
  const [addresses, setAddresses] = useState([])
  const [modalType, setModalType] = useState('')
  const [selectedAsset, setSelectedAsset] = useState(null)
  const [createForm] = Form.useForm()
  const [mintForm] = Form.useForm()

  const fetchAssetsAndAddresses = async () => {
    const [fetchedAssets, fetchedAddresses] = await Promise.all([
      GetMyAssets(),
      GetAddressBook()
    ])
    setAssets(fetchedAssets)
    setAddresses(
      fetchedAddresses.map((addr) => ({
        label: `${addr.Name}: ${addr.Address}`,
        value: addr.Address
      }))
    )
  }

  useEffect(() => {
    fetchAssetsAndAddresses()
  }, [])

  const handleCreateAsset = async (values) => {
    setModalType('') // Close the modal immediately
    message.loading({
      content: 'Processing Transaction...',
      key: 'createProcess',
      duration: 0
    })
    try {
      const start = new Date().getTime()
      await CreateAsset(values.symbol, `${values.decimals}`, values.metadata)
      const finish = new Date().getTime()
      let timeTaken = (finish - start) / 1000 // Time taken in seconds
      timeTaken =
        timeTaken < 1 ? 'less than a second' : `${timeTaken.toFixed(1)} seconds`
      message.success({
        content: `Token created successfully in ${timeTaken}`,
        key: 'createProcess',
        duration: 5
      })
      fetchAssetsAndAddresses()
      createForm.resetFields() // Reset form fields
    } catch (error) {
      message.error({
        content: `Failed to create token: ${error}`,
        key: 'createProcess',
        duration: 5
      })
    }
  }

  const handleMintAsset = async (values) => {
    setModalType('') // Close the mint modal immediately by resetting modalType
    message.loading({
      content: 'Processing Transaction...',
      key: 'mintProcess',
      duration: 0
    })
    try {
      const start = new Date().getTime()
      await MintAsset(
        selectedAsset.ID,
        values.address,
        values.amount.toString()
      )
      const finish = new Date().getTime()
      let timeTaken = (finish - start) / 1000 // Time taken in seconds
      timeTaken =
        timeTaken < 1 ? 'less than a second' : `${timeTaken.toFixed(1)} seconds`
      message.success({
        content: `Token minted successfully in ${timeTaken}`,
        key: 'mintProcess',
        duration: 5
      })
      fetchAssetsAndAddresses() // Refetch assets and addresses to update the UI
      mintForm.resetFields() // Reset form fields after minting
    } catch (error) {
      message.error({
        content: `Failed to mint token: ${error}`,
        key: 'mintProcess',
        duration: 5
      })
    }
  }

  return (
    <div style={{ maxWidth: '700px', margin: 'auto', padding: '20px' }}>
      <FundsCheck />
      <Card>
        <Space style={{ width: '100%', justifyContent: 'space-between' }}>
          <Title level={4}>Your Tokens</Title>
          <Button onClick={() => setModalType('create')}>Create Token</Button>
        </Space>
        {assets.map((asset) => (
          <Card
            type='inner'
            key={asset.ID}
            title={asset.Symbol}
            extra={
              <Button
                onClick={() => {
                  setSelectedAsset(asset)
                  setModalType('mint')
                }}
              >
                Mint
              </Button>
            }
          >
            <p>Decimals: {asset.Decimals}</p>
            <p>Supply: {asset.Supply}</p>
          </Card>
        ))}
      </Card>

      <Modal
        title='Create New Token'
        open={modalType === 'create'}
        onCancel={() => setModalType('')}
        footer={null}
      >
        <Form form={createForm} layout='vertical' onFinish={handleCreateAsset}>
          <Form.Item name='symbol' label='Symbol' rules={[{ required: true }]}>
            <Input placeholder='e.g. ABC' />
          </Form.Item>
          <Form.Item
            name='decimals'
            label='Decimals'
            rules={[{ required: true }]}
          >
            <InputNumber
              min={0}
              max={18}
              placeholder='e.g. 8'
              style={{ width: '100%' }}
            />
          </Form.Item>
          <Form.Item name='metadata' label='Metadata (optional)'>
            <Input.TextArea rows={4} />
          </Form.Item>
          <Form.Item>
            <Button type='primary' htmlType='submit'>
              Create
            </Button>
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title={`Mint ${selectedAsset?.Symbol}`}
        open={modalType === 'mint'}
        onCancel={() => setModalType('')}
        footer={null}
      >
        <Form layout='vertical' form={mintForm} onFinish={handleMintAsset}>
          <Form.Item
            name='address'
            label='Recipient Address'
            rules={[{ required: true }]}
          >
            <Select options={addresses} placeholder='Select recipient' />
          </Form.Item>
          <Form.Item name='amount' label='Amount' rules={[{ required: true }]}>
            <InputNumber
              style={{ width: '100%' }}
              min={0}
              placeholder='e.g. 100'
            />
          </Form.Item>
          <Form.Item>
            <Button type='primary' htmlType='submit'>
              Mint
            </Button>
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default Mint
