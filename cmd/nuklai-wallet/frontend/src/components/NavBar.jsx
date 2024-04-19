// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import {
  BankOutlined,
  CheckCircleTwoTone,
  CloseCircleTwoTone,
  ContainerOutlined,
  CopyOutlined,
  DashboardOutlined,
  GoldOutlined,
  SendOutlined,
  WalletOutlined,
  WalletTwoTone
} from '@ant-design/icons'
import {
  App,
  Avatar,
  Button,
  Card,
  Divider,
  Drawer,
  Form,
  Input,
  Layout,
  List,
  Menu,
  Tooltip,
  Typography
} from 'antd'
import { useCallback, useEffect, useRef, useState } from 'react'
import { Link as RLink, useLocation } from 'react-router-dom'
import {
  GetAddress,
  GetBalance,
  GetChainID,
  GetConfig,
  GetPrivateKey,
  GetPublicKey,
  GetSubnetID,
  GetTransactions,
  UpdateFaucetRPC,
  UpdateFeedRPC,
  UpdateNuklaiRPC
} from '../../wailsjs/go/main/App'
import logo from '../assets/images/nuklai-logo.png'
const { Text, Link } = Typography

const NavBar = () => {
  const location = useLocation()
  const { message } = App.useApp()
  const [nuklaiRPC, setNuklaiRPC] = useState('')
  const [faucetRPC, setFaucetRPC] = useState('')
  const [feedRPC, setFeedRPC] = useState('')
  const [balances, setBalances] = useState([])
  const [nativeBalance, setNativeBalance] = useState({})
  const [transactions, setTransactions] = useState([])
  const [address, setAddress] = useState('')
  const [privateKey, setPrivateKey] = useState('')
  const [publicKey, setPublicKey] = useState('')
  const [subnetID, setSubnetID] = useState('')
  const [chainID, setChainID] = useState('')
  const [form] = Form.useForm() // Form layout for better alignment and submission handling
  const [drawerOpen, setDrawerOpen] = useState(false)

  const isMounted = useRef(true) // Track if component is mounted

  useEffect(() => {
    return () => {
      isMounted.current = false // Set to false when component unmounts
    }
  }, [])

  // Define the function to fetch data as a useCallback to prevent redefinition on each render
  const fetchData = useCallback(async () => {
    try {
      const [newAddress, newBalances, txs] = await Promise.all([
        GetAddress(),
        GetBalance(),
        GetTransactions()
      ])

      setAddress(newAddress)
      setBalances(newBalances)
      setTransactions(txs.TxInfos)

      for (var bal of newBalances) {
        if (bal.ID == '11111111111111111111111111111111LpoYY') {
          setNativeBalance(bal)
          {
            /* TODO: switch to using context */
          }
          window.HasBalance = bal.Has
          break
        }
      }

      // Handle alerts
      txs.Alerts?.forEach((alert) =>
        message.open({
          icon: <WalletTwoTone />,
          type: alert.Type,
          content: alert.Content
        })
      )
    } catch (error) {
      console.error('Error fetching data:', error)
      message.error('Failed to fetch data. Please try again later.')
    }
  }, [message])

  useEffect(() => {
    fetchData()
    fetchWalletInfo()
    const intervalId = setInterval(fetchData, 5000) // Update every 5 seconds
    return () => clearInterval(intervalId)
  }, [fetchData])

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  const shortenText = (text, cutoffIndex) =>
    `${text.slice(0, cutoffIndex)}...${text.slice(-cutoffIndex)}`

  // Fetch wallet info
  const fetchWalletInfo = async () => {
    try {
      const [
        currentConfig,
        fetchedPrivateKey,
        fetchedPublicKey,
        fetchedSubnetID,
        fetchedChainID
      ] = await Promise.all([
        GetConfig(),
        GetPrivateKey(),
        GetPublicKey(),
        GetSubnetID(),
        GetChainID()
      ])

      setNuklaiRPC(currentConfig.nuklaiRPC)
      setFaucetRPC(currentConfig.faucetRPC)
      setFeedRPC(currentConfig.feedRPC)
      setPrivateKey(fetchedPrivateKey)
      setPublicKey(fetchedPublicKey)
      setSubnetID(fetchedSubnetID)
      setChainID(fetchedChainID)

      // Update form values
      if (isMounted.current) {
        form.setFieldsValue({
          nuklaiRPC: currentConfig.nuklaiRPC,
          faucetRPC: currentConfig.faucetRPC,
          feedRPC: currentConfig.feedRPC
        })
      }
    } catch (error) {
      console.error('Error fetching wallet info:', error)
    }
  }

  const handleUpdateRPC = async (field) => {
    const value = form.getFieldValue(field)
    try {
      switch (field) {
        case 'nuklaiRPC':
          await UpdateNuklaiRPC(value)
          message.success('NuklaiRPC updated successfully')
          break
        case 'faucetRPC':
          await UpdateFaucetRPC(value)
          message.success('FaucetRPC updated successfully')
          break
        case 'feedRPC':
          await UpdateFeedRPC(value)
          message.success('FeedRPC updated successfully')
          break
        default:
          break
      }
      fetchData()
      fetchWalletInfo()

      // Update form values to ensure UI consistency
      form.setFieldsValue({ [field]: value })
    } catch (error) {
      console.error(`Failed to update ${field}:`, error)
      message.error(`Failed to update ${field}. Please try again.`)
    }
  }

  // Function to handle the copying and showing the notification
  const handleCopyTokenId = (tokenId) => {
    navigator.clipboard.writeText(tokenId).then(
      () => {
        message.success('Token ID copied to clipboard')
      },
      (err) => {
        message.error('Failed to copy Token ID')
      }
    )
  }

  return (
    <Layout.Header style={{ padding: '0 50px', background: '#fff' }}>
      <div className='logo' style={{ float: 'left' }}>
        <Avatar src={logo} size='large' />
      </div>

      <div
        className='walletInfo'
        style={{ float: 'right', display: 'flex', alignItems: 'center' }}
      >
        <Avatar
          icon={<WalletOutlined />}
          onClick={() => setDrawerOpen(true)}
          style={{ cursor: 'pointer', marginRight: '10px' }}
        />
        <div style={{ display: 'flex', flexDirection: 'column' }}>
          <Tooltip title={address}>
            <Link
              strong
              onClick={() => setDrawerOpen(true)}
              style={{ marginTop: '10px' }}
            >
              {shortenText(address, 7)}
            </Link>
          </Tooltip>
          {balances.length > 0 && (
            <Link
              strong
              onClick={() => setDrawerOpen(true)}
              style={{ marginTop: '5px' }}
            >
              {nativeBalance.Str}
            </Link>
          )}
        </div>
      </div>

      <Menu
        mode='horizontal'
        selectedKeys={[location.pathname.slice(1) || 'explorer']}
        items={[
          {
            label: <RLink to={'explorer'}>Explorer</RLink>,
            key: 'explorer',
            icon: <DashboardOutlined />
          },
          {
            label: <RLink to={'faucet'}>Faucet</RLink>,
            key: 'faucet',
            icon: <GoldOutlined />
          },
          {
            label: <RLink to={'mint'}>Mint</RLink>,
            key: 'mint',
            icon: <BankOutlined />
          },
          {
            label: <RLink to={'transfer'}>Transfer</RLink>,
            key: 'transfer',
            icon: <SendOutlined />
          },
          {
            label: <RLink to={'feed'}>Feed</RLink>,
            key: 'feed',
            icon: <ContainerOutlined />
          }
        ]}
      />

      <Drawer
        title={<Text copyable>{address}</Text>}
        width={615}
        placement='right'
        onClose={() => setDrawerOpen(false)}
        open={drawerOpen}
      >
        <Divider orientation='center'>Wallet Info</Divider>
        <div style={{ marginBottom: '20px' }}>
          <List
            itemLayout='horizontal'
            dataSource={[
              { label: 'Private Key', value: privateKey, key: 'privateKey' },
              { label: 'Public Key', value: publicKey, key: 'publicKey' },
              { label: 'Subnet ID', value: subnetID, key: 'subnetID' },
              { label: 'Chain ID', value: chainID, key: 'chainID' }
            ]}
            renderItem={(item) => (
              <List.Item>
                <List.Item.Meta
                  title={item.label}
                  description={
                    <Tooltip title={item.value}>
                      <Text
                        code
                        copyable={{ text: item.value }}
                        style={{
                          background: '#f5f5f5',
                          padding: '4px 8px',
                          borderRadius: '4px',
                          maxWidth: '100%',
                          overflow: 'hidden',
                          textOverflow: 'ellipsis',
                          whiteSpace: 'nowrap',
                          cursor: 'pointer'
                        }}
                      >
                        {item.key !== 'chainID' && item.key !== 'subnetID'
                          ? shortenText(item.value, 35)
                          : item.value}
                      </Text>
                    </Tooltip>
                  }
                />
              </List.Item>
            )}
          />
        </div>

        <Card
          title='RPC Configuration'
          bordered={false}
          style={{ margin: '24px' }}
        >
          <Form
            form={form} // Make sure this is correctly passed
            layout='vertical'
            initialValues={{
              nuklaiRPC: nuklaiRPC,
              faucetRPC: faucetRPC,
              feedRPC: feedRPC
            }}
          >
            <Form.Item label='NuklaiRPC URL' name='nuklaiRPC'>
              <Input
                addonAfter={
                  <Button onClick={() => handleUpdateRPC('nuklaiRPC')}>
                    Apply
                  </Button>
                }
              />
            </Form.Item>
            <Form.Item label='FaucetRPC URL' name='faucetRPC'>
              <Input
                addonAfter={
                  <Button onClick={() => handleUpdateRPC('faucetRPC')}>
                    Apply
                  </Button>
                }
              />
            </Form.Item>
            <Form.Item label='FeedRPC URL' name='feedRPC'>
              <Input
                addonAfter={
                  <Button onClick={() => handleUpdateRPC('feedRPC')}>
                    Apply
                  </Button>
                }
              />
            </Form.Item>
          </Form>
        </Card>

        <Divider orientation='center'>Tokens</Divider>
        <List
          bordered
          dataSource={balances}
          renderItem={(balance) => {
            // Extract token information and optional token ID
            const tokenRegex = /^(.+)\s\[(.+)\]$/ // Regex to extract token details and ID
            const match = balance.Str.match(tokenRegex)
            let tokenDisplay, tokenId

            if (match) {
              tokenDisplay = match[1]
              tokenId = match[2]
            } else {
              tokenDisplay = balance.Str
              tokenId = ''
            }

            return (
              <List.Item
                style={{
                  display: 'flex',
                  justifyContent: 'space-between',
                  alignItems: 'center',
                  padding: '10px'
                }}
              >
                <Text strong>{tokenDisplay}</Text>
                {tokenId ? (
                  <Button
                    type='text'
                    icon={<CopyOutlined />}
                    onClick={() => handleCopyTokenId(tokenId)}
                    style={{ color: '#1890ff' }} // Color used to indicate interactivity
                  >
                    [{tokenId}]
                  </Button>
                ) : (
                  <Text type='secondary'>Native Asset</Text>
                )}
              </List.Item>
            )
          }}
          style={{
            background: '#f0f2f5',
            borderRadius: '4px',
            overflow: 'hidden'
          }}
        />

        <Divider orientation='center'>Transactions</Divider>
        <List
          bordered
          dataSource={transactions}
          renderItem={(item) => (
            <List.Item>
              <div>
                <Text strong>{item.ID} </Text>
                {!item.Success && <CloseCircleTwoTone twoToneColor='#eb2f96' />}
                {item.Success && <CheckCircleTwoTone twoToneColor='#52c41a' />}
              </div>
              <Text strong>Type:</Text> {item.Type}
              <br />
              <Text strong>Timestamp:</Text> {formatTimestamp(item.Timestamp)}
              <br />
              <Text strong>Units:</Text> {item.Units}
              <br />
              <Text strong>Size:</Text> {item.Size}
              <br />
              <Text strong>Summary:</Text> {item.Summary}
              <br />
              <Text strong>Fee:</Text> {item.Fee}
              <br />
              <Text strong>Actor:</Text>{' '}
              <Text copyable={{ text: item.Actor }}>
                {shortenText(item.Actor, 20)}
              </Text>
            </List.Item>
          )}
        />
      </Drawer>
    </Layout.Header>
  )
}

export default NavBar
