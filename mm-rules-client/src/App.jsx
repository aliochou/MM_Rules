import React, { useState } from 'react'
import reactLogo from './assets/react.svg'
import viteLogo from '/vite.svg'
import './App.css'

const BACKEND_URL = 'http://localhost:8080'

// Game configurations
const GAME_MODES = {
  '1v1': {
    id: 'game-1v1',
    name: '1v1 Competitive',
    description: 'Head-to-head competitive matches',
    teams: ['Player1', 'Player2'],
    teamSize: 1
  },
  '1v3': {
    id: 'game-1v3',
    name: '1v3 Team Battle',
    description: 'Solo player vs trio team',
    teams: ['Solo', 'Trio'],
    teamSize: { Solo: 1, Trio: 3 }
  }
}

function App() {
  const [status, setStatus] = useState('')
  const [requestId, setRequestId] = useState(null)
  const [session, setSession] = useState(null)
  const [matchInfo, setMatchInfo] = useState(null)
  const [loading, setLoading] = useState(false)
  const [selectedGameMode, setSelectedGameMode] = useState(null)
  const [currentPlayerId, setCurrentPlayerId] = useState(null)

  const generatePlayerMetadata = (gameMode) => {
    const baseMetadata = {
      level: Math.floor(Math.random() * 50) + 10, // Random level 10-60
      region: 'us-west',
    }

    if (gameMode === '1v1') {
      return {
        ...baseMetadata,
        skill_rating: Math.floor(Math.random() * 1000) + 1000, // 1000-2000
        preferred_role: ['attacker', 'defender'][Math.floor(Math.random() * 2)]
      }
    } else if (gameMode === '1v3') {
      return {
        ...baseMetadata,
        team_experience: Math.floor(Math.random() * 5) + 1, // 1-5
        communication: ['voice', 'text'],
        preferred_role: ['leader', 'support', 'attacker', 'defender'][Math.floor(Math.random() * 4)]
      }
    }

    return baseMetadata
  }

  const joinGameMode = async (gameMode) => {
    const gameConfig = GAME_MODES[gameMode]
    if (!gameConfig) {
      setStatus('Invalid game mode')
      return
    }

    const playerId = `player-${gameMode}-${Math.floor(Math.random() * 100000)}`
    setStatus(`Joining ${gameConfig.name}...`)
    setSession(null)
    setMatchInfo(null)
    setLoading(true)
    setRequestId(null)
    setSelectedGameMode(gameMode)
    setCurrentPlayerId(playerId)

    try {
      // Step 1: Create match request
      const res = await fetch(`${BACKEND_URL}/api/v1/match-request`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ 
          player_id: playerId, 
          game_id: gameConfig.id, 
          metadata: generatePlayerMetadata(gameMode)
        }),
      })
      
      const data = await res.json()
      if (!data.request_id) throw new Error('No request_id returned')
      setRequestId(data.request_id)
      setStatus(`Waiting for ${gameConfig.name} match...`)
      
      // Step 2: Trigger matchmaking processing
      const matchmakingRes = await fetch(`${BACKEND_URL}/api/v1/process-matchmaking/${gameConfig.id}`, { 
        method: 'POST' 
      })
      const matchmakingData = await matchmakingRes.json()
      
      // Find the match that contains our player
      if (matchmakingData.matches && matchmakingData.matches.length > 0) {
        const playerMatch = matchmakingData.matches.find(match => 
          match.players && match.players.includes(playerId)
        )
        
        if (playerMatch) {
          setMatchInfo({
            matchId: playerMatch.match_id,
            players: playerMatch.players,
            teamName: playerMatch.team_name,
            createdAt: playerMatch.created_at,
            gameMode: gameMode,
            playerId: playerId
          })
          setStatus(`Match found! ${gameConfig.name}`)
        } else {
          setStatus('Match created but player not found in any match')
        }
      }
      
      // Step 3: Poll for match status
      pollStatus(data.request_id, 0, gameConfig.id)
    } catch (err) {
      setStatus('Error: ' + err.message)
      setLoading(false)
    }
  }

  const pollStatus = async (reqId, attempt, gameId) => {
    if (attempt > 30) {
      setStatus('Timed out waiting for match.')
      setLoading(false)
      return
    }
    try {
      const res = await fetch(`${BACKEND_URL}/api/v1/match-status/${reqId}`)
      const data = await res.json()
      if (data.status === 'matched' || data.status === 'allocated') {
        // Always fetch match info from the backend when matched
        try {
          const matchmakingRes = await fetch(`${BACKEND_URL}/api/v1/process-matchmaking/${gameId}`, { 
            method: 'POST' 
          })
          const matchmakingData = await matchmakingRes.json()
          // Find the match that contains our player
          if (matchmakingData.matches && matchmakingData.matches.length > 0) {
            const playerMatch = matchmakingData.matches.find(match => 
              match.players && match.players.includes(currentPlayerId)
            )
            if (playerMatch) {
              setMatchInfo({
                matchId: playerMatch.match_id,
                players: playerMatch.players,
                teamName: playerMatch.team_name,
                createdAt: playerMatch.created_at,
                gameMode: selectedGameMode,
                playerId: currentPlayerId
              })
            }
          }
        } catch (err) {
          console.log('Could not fetch match details:', err)
        }
        setStatus('Match found!')
        if (data.session) {
          setSession(data.session)
          setStatus('Match is running!')
        }
        setLoading(false)
        return
      }
      setTimeout(() => pollStatus(reqId, attempt + 1, gameId), 1000)
    } catch (err) {
      setStatus('Error polling status: ' + err.message)
      setLoading(false)
    }
  }

  const getTeamDisplayName = (teamName, gameMode) => {
    if (gameMode === '1v1') {
      return teamName === 'Player1' ? 'Team A' : 'Team B'
    } else if (gameMode === '1v3') {
      return teamName === 'Solo' ? 'Solo Player' : 'Trio Team'
    }
    return teamName
  }

  const getTeamSize = (teamName, gameMode) => {
    if (gameMode === '1v1') {
      return 1
    } else if (gameMode === '1v3') {
      return teamName === 'Solo' ? 1 : 3
    }
    return 1
  }

  return (
    <div style={{ 
      display: 'flex', 
      flexDirection: 'column', 
      alignItems: 'center', 
      marginTop: 40,
      fontFamily: 'Arial, sans-serif',
      maxWidth: '800px',
      margin: '40px auto',
      padding: '0 20px'
    }}>
      <h1 style={{ 
        color: '#2c3e50', 
        marginBottom: '10px',
        textAlign: 'center'
      }}>
        MM-Rules Matchmaking
      </h1>
      <p style={{ 
        color: '#7f8c8d', 
        marginBottom: '40px',
        textAlign: 'center',
        fontSize: '18px'
      }}>
        Choose your game mode and join the queue
      </p>

      {/* Game Mode Selection */}
      <div style={{ 
        display: 'flex', 
        gap: '20px', 
        marginBottom: '40px',
        flexWrap: 'wrap',
        justifyContent: 'center'
      }}>
        {Object.entries(GAME_MODES).map(([mode, config]) => (
          <div key={mode} style={{
            border: '2px solid #3498db',
            borderRadius: '12px',
            padding: '24px',
            textAlign: 'center',
            minWidth: '200px',
            backgroundColor: '#f8f9fa',
            transition: 'all 0.3s ease',
            cursor: loading ? 'not-allowed' : 'pointer',
            opacity: loading ? 0.6 : 1,
            transform: loading ? 'scale(0.98)' : 'scale(1)'
          }}>
            <h3 style={{ 
              margin: '0 0 8px 0', 
              color: '#2c3e50',
              fontSize: '20px'
            }}>
              {config.name}
            </h3>
            <p style={{ 
              margin: '0 0 16px 0', 
              color: '#7f8c8d',
              fontSize: '14px'
            }}>
              {config.description}
            </p>
            <button 
              onClick={() => joinGameMode(mode)}
              disabled={loading}
              style={{
                fontSize: '16px',
                padding: '12px 24px',
                backgroundColor: '#3498db',
                color: 'white',
                border: 'none',
                borderRadius: '8px',
                cursor: loading ? 'not-allowed' : 'pointer',
                transition: 'background-color 0.3s ease',
                fontWeight: 'bold'
              }}
              onMouseOver={(e) => {
                if (!loading) e.target.style.backgroundColor = '#2980b9'
              }}
              onMouseOut={(e) => {
                if (!loading) e.target.style.backgroundColor = '#3498db'
              }}
            >
              {loading && selectedGameMode === mode ? 'Joining...' : `Join ${config.name}`}
            </button>
          </div>
        ))}
      </div>

      {/* Status Display */}
      {status && (
        <div style={{ 
          marginBottom: '20px', 
          fontSize: '18px',
          padding: '16px',
          backgroundColor: '#e8f4fd',
          borderRadius: '8px',
          border: '1px solid #3498db',
          color: '#2c3e50',
          textAlign: 'center',
          minWidth: '400px'
        }}>
          {status}
        </div>
      )}
      
      {/* Match Information */}
      {matchInfo && (
        <div style={{ 
          marginTop: '20px', 
          fontSize: '16px', 
          padding: '24px', 
          border: '2px solid #27ae60', 
          borderRadius: '12px',
          backgroundColor: '#f8fff9',
          minWidth: '500px',
          boxShadow: '0 4px 6px rgba(0,0,0,0.1)'
        }}>
          <h3 style={{ 
            margin: '0 0 16px 0', 
            color: '#27ae60',
            textAlign: 'center',
            fontSize: '20px'
          }}>
            🎮 Match Found!
          </h3>
          
          <div style={{ display: 'grid', gap: '12px' }}>
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid #e0e0e0'
            }}>
              <strong>Match ID:</strong>
              <span style={{ fontFamily: 'monospace', color: '#2c3e50' }}>
                {matchInfo.matchId}
              </span>
            </div>
            
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid #e0e0e0'
            }}>
              <strong>Game Mode:</strong>
              <span style={{ color: '#3498db', fontWeight: 'bold' }}>
                {GAME_MODES[matchInfo.gameMode].name}
              </span>
            </div>
            
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid #e0e0e0'
            }}>
              <strong>Your Team:</strong>
              <span style={{ 
                color: '#e74c3c', 
                fontWeight: 'bold',
                backgroundColor: '#fdf2f2',
                padding: '4px 8px',
                borderRadius: '4px'
              }}>
                {getTeamDisplayName(matchInfo.teamName, matchInfo.gameMode)}
              </span>
            </div>
            
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid #e0e0e0'
            }}>
              <strong>Your Player ID:</strong>
              <span style={{ 
                fontFamily: 'monospace', 
                color: '#2c3e50',
                fontSize: '14px'
              }}>
                {matchInfo.playerId}
              </span>
            </div>
            
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid #e0e0e0'
            }}>
              <strong>Team Size:</strong>
              <span style={{ color: '#2c3e50' }}>
                {getTeamSize(matchInfo.teamName, matchInfo.gameMode)} player(s)
              </span>
            </div>
            
            <div style={{ 
              display: 'flex', 
              justifyContent: 'space-between',
              padding: '8px 0',
              borderBottom: '1px solid #e0e0e0'
            }}>
              <strong>Total Players:</strong>
              <span style={{ color: '#2c3e50' }}>
                {matchInfo.players.length} player(s)
              </span>
            </div>
          </div>
          
          <div style={{ 
            marginTop: '16px',
            padding: '12px',
            backgroundColor: '#f0f8ff',
            borderRadius: '8px',
            border: '1px solid #3498db'
          }}>
            <strong style={{ color: '#3498db', display: 'block', marginBottom: '8px' }}>
              All Players in Match:
            </strong>
            <div style={{ 
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fit, minmax(200px, 1fr))',
              gap: '8px'
            }}>
              {matchInfo.players.map((player, index) => (
                <div key={index} style={{ 
                  padding: '8px 12px',
                  backgroundColor: player === matchInfo.playerId ? '#e8f5e8' : '#f8f9fa',
                  borderRadius: '6px',
                  fontSize: '16px',
                  fontFamily: 'Arial, sans-serif',
                  fontWeight: '500',
                  border: player === matchInfo.playerId ? '2px solid #27ae60' : '1px solid #e0e0e0',
                  color: '#2c3e50',
                  textAlign: 'center'
                }}>
                  {player === matchInfo.playerId ? '👤 ' : '• '}{player}
                </div>
              ))}
            </div>
          </div>
          
          <div style={{ 
            marginTop: '12px', 
            fontSize: '14px', 
            color: '#7f8c8d',
            textAlign: 'center',
            fontStyle: 'italic',
            fontFamily: 'Arial, sans-serif'
          }}>
            Created: {new Date(matchInfo.createdAt).toLocaleString()}
          </div>
        </div>
      )}
      
      {/* Session Information */}
      {session && (
        <div style={{ 
          marginTop: '20px', 
          fontSize: '16px',
          padding: '20px',
          border: '2px solid #f39c12',
          borderRadius: '12px',
          backgroundColor: '#fef9e7',
          minWidth: '400px',
          textAlign: 'center'
        }}>
          <h3 style={{ 
            margin: '0 0 16px 0', 
            color: '#f39c12',
            fontSize: '18px'
          }}>
            🎯 Game Session Active
          </h3>
          <div style={{ display: 'grid', gap: '8px' }}>
            <div><strong>Server IP:</strong> {session.ip}</div>
            <div><strong>Port:</strong> {session.port}</div>
            <div><strong>Session ID:</strong> {session.id}</div>
          </div>
        </div>
      )}
    </div>
  )
}

export default App
