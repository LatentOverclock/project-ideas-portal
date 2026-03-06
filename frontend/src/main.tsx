import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'

type Idea = { id: string; title: string; description: string; userEmail: string; createdAt: string }
const API_URL = (import.meta.env.VITE_API_URL as string) || '/graphql'

async function gql(query: string, variables: Record<string, unknown> = {}, token?: string) {
  const res = await fetch(API_URL, {
    method: 'POST', headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}) },
    body: JSON.stringify({ query, variables })
  })
  const data = await res.json()
  if (data.errors?.length) throw new Error(data.errors[0].message)
  return data.data
}

function App() {
  const [token, setToken] = useState(localStorage.getItem('token') || '')
  const [email, setEmail] = useState(localStorage.getItem('email') || '')
  const [ideas, setIdeas] = useState<Idea[]>([])
  const [authEmail, setAuthEmail] = useState('')
  const [password, setPassword] = useState('')
  const [title, setTitle] = useState('')
  const [description, setDescription] = useState('')
  const [error, setError] = useState('')

  const loadIdeas = async () => { const d = await gql(`query { ideas { id title description userEmail createdAt } }`); setIdeas(d.ideas) }
  useEffect(() => { loadIdeas().catch((e) => setError(String(e.message || e))) }, [])

  const handleAuth = async (mode: 'register' | 'login') => {
    setError('')
    const mutation = mode === 'register'
      ? `mutation($email:String!,$password:String!){ register(email:$email,password:$password){ token email } }`
      : `mutation($email:String!,$password:String!){ login(email:$email,password:$password){ token email } }`
    const data = await gql(mutation, { email: authEmail, password })
    const payload = mode === 'register' ? data.register : data.login
    setToken(payload.token); setEmail(payload.email)
    localStorage.setItem('token', payload.token); localStorage.setItem('email', payload.email)
  }

  const submitIdea = async () => {
    setError('')
    await gql(`mutation($title:String!,$description:String!){ createIdea(title:$title,description:$description){ id } }`, { title, description }, token)
    setTitle(''); setDescription(''); await loadIdeas()
  }

  const logout = () => { setToken(''); setEmail(''); localStorage.clear() }

  return <main style={{ maxWidth: 900, margin: '2rem auto', fontFamily: 'system-ui', padding: '0 1rem' }}>
    <h1>Project Ideas Portal</h1>
    <p>Register/login, then submit ideas that can later be implemented by an agent.</p>
    {error && <p style={{ color: 'crimson' }}>{error}</p>}

    {!token ? <section>
      <h2>Account</h2>
      <input placeholder="Email" value={authEmail} onChange={e => setAuthEmail(e.target.value)} />{' '}
      <input placeholder="Password (min 8 chars)" type="password" value={password} onChange={e => setPassword(e.target.value)} />{' '}
      <button onClick={() => handleAuth('register')}>Register</button>{' '}
      <button onClick={() => handleAuth('login')}>Login</button>
    </section> : <section>
      <p>Logged in as <strong>{email}</strong> <button onClick={logout}>Logout</button></p>
      <h2>Submit idea</h2>
      <input placeholder="Title" value={title} onChange={e => setTitle(e.target.value)} style={{ width: '100%', marginBottom: 8 }} />
      <textarea placeholder="Description" value={description} onChange={e => setDescription(e.target.value)} style={{ width: '100%', minHeight: 100 }} />
      <br /><button onClick={submitIdea}>Create idea</button>
    </section>}

    <section>
      <h2>Ideas</h2>
      <ul>
        {ideas.map(i => <li key={i.id} style={{ marginBottom: 12 }}><strong>{i.title}</strong> — by {i.userEmail}<br />{i.description}<br /><small>{new Date(i.createdAt).toLocaleString()}</small></li>)}
      </ul>
    </section>
  </main>
}

createRoot(document.getElementById('root')!).render(<App />)
