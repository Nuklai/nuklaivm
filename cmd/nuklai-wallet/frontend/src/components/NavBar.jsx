// Copyright (C) 2024, AllianceBlock. All rights reserved.
// See the file LICENSE for licensing terms.

import {
  BankOutlined,
  CheckCircleTwoTone,
  CloseCircleTwoTone,
  ContainerOutlined,
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
  Divider,
  Drawer,
  Input,
  Layout,
  List,
  Menu,
  Tooltip,
  Typography
} from 'antd'
import { useCallback, useEffect, useState } from 'react'
import { Link as RLink, useLocation } from 'react-router-dom'
import {
  GetAddress,
  GetBalance,
  GetChainID,
  GetConfig,
  GetPrivateKey,
  GetPublicKey,
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
  const [chainID, setChainID] = useState('')
  const [drawerOpen, setDrawerOpen] = useState(false)

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
    const intervalId = setInterval(fetchData, 5000) // Update every 5 seconds
    return () => clearInterval(intervalId)
  }, [fetchData])

  const formatTimestamp = (timestamp) => {
    const date = new Date(timestamp)
    return date.toLocaleString()
  }

  const shortenText = (text, cutoffIndex) =>
    `${text.slice(0, cutoffIndex)}...${text.slice(-cutoffIndex)}`

  // Fetch privateKey and chainID (this is a simplified example, adjust according to your actual API)
  useEffect(() => {
    // Fetch private key and chainID securely
    const fetchWalletInfo = async () => {
      try {
        const fetchedPrivateKey = await GetPrivateKey()
        const fetchedPublicKey = await GetPublicKey()
        const fetchedChainID = await GetChainID()

        setPrivateKey(fetchedPrivateKey)
        setPublicKey(fetchedPublicKey)
        setChainID(fetchedChainID)
      } catch (error) {
        console.error('Error fetching wallet info:', error)
      }
    }

    fetchWalletInfo()
  }, [])

  const fetchConfig = async () => {
    const currentConfig = await GetConfig()
    setNuklaiRPC(currentConfig.nuklaiRPC)
    setFaucetRPC(currentConfig.faucetRPC)
    setFeedRPC(currentConfig.feedRPC)
  }

  useEffect(() => {
    fetchConfig()
  }, [])

  const updateNuklaiRPC = async () => {
    await UpdateNuklaiRPC(nuklaiRPC)
    message.success('NuklaiRPC updated successfully')
  }

  const updateFaucetRPC = async () => {
    await UpdateFaucetRPC(faucetRPC)
    message.success('FaucetRPC updated successfully')
  }

  const updateFeedRPC = async () => {
    await UpdateFeedRPC(feedRPC)
    message.success('FeedRPC updated successfully')
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
                        {item.key !== 'chainID'
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

        <Divider orientation='center'>RPC Configuration</Divider>
        <div>
          <Input
            addonBefore='NuklaiRPC'
            value={nuklaiRPC}
            onChange={(e) => setNuklaiRPC(e.target.value)}
          />
          <Button onClick={updateNuklaiRPC} style={{ margin: '10px 0' }}>
            Update NuklaiRPC
          </Button>
        </div>
        <div>
          <Input
            addonBefore='FaucetRPC'
            value={faucetRPC}
            onChange={(e) => setFaucetRPC(e.target.value)}
          />
          <Button onClick={updateFaucetRPC} style={{ margin: '10px 0' }}>
            Update FaucetRPC
          </Button>
        </div>
        <div>
          <Input
            addonBefore='FeedRPC'
            value={feedRPC}
            onChange={(e) => setFeedRPC(e.target.value)}
          />
          <Button onClick={updateFeedRPC} style={{ margin: '10px 0' }}>
            Update FeedRPC
          </Button>
        </div>

        <Divider orientation='center'>Tokens</Divider>
        <List
          bordered
          dataSource={balances}
          renderItem={(balance) => (
            <List.Item>
              <Text>{balance.Str}</Text>
            </List.Item>
          )}
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
              <Text copyable>{shortenText(item.Actor, 20)}</Text>
            </List.Item>
          )}
        />
      </Drawer>
    </Layout.Header>
  )
}

export default NavBar
