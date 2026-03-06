import React, { useEffect, useState } from 'react'
import { createRoot } from 'react-dom/client'
import './styles.css'

type Idea = { id: string; title: string; description: string; userEmail: string; createdAt: string }
const API_URL = '/graphql'

async function gql(query: string, variables: Record<string, unknown> = {}, token?: string) {
  const response = await fetch(API_URL, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {})
    },
    body: JSON.stringify({ query, variables })
  })
  const payload = await response.json()
  if (payload.errors?.length) throw new Error(payload.errors[0].message)
  return payload.data
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

  const loadIdeas = async () => {
    const data = await gql(`query { ideas { id title description userEmail createdAt } }`)
    setIdeas(data.ideas)
  }

  useEffect(() => {
    loadIdeas().catch((e) => setError(String(e.message || e)))
  }, [])

  const handleAuth = async (mode: 'register' | 'login') => {
    setError('')
    const mutation =
      mode === 'register'
        ? `mutation($email:String!,$password:String!){ register(email:$email,password:$password){ token email } }`
        : `mutation($email:String!,$password:String!){ login(email:$email,password:$password){ token email } }`
    const data = await gql(mutation, { email: authEmail, password })
    const auth = mode === 'register' ? data.register : data.login
    setToken(auth.token)
    setEmail(auth.email)
    localStorage.setItem('token', auth.token)
    localStorage.setItem('email', auth.email)
  }

  const submitIdea = async () => {
    setError('')
    await gql(
      `mutation($title:String!,$description:String!){ createIdea(title:$title,description:$description){ id } }`,
      { title, description },
      token,
    )
    setTitle('')
    setDescription('')
    await loadIdeas()
  }

  const logout = () => {
    setToken('')
    setEmail('')
    localStorage.removeItem('token')
    localStorage.removeItem('email')
  }

  return (
    <main className="page">
      <h1>Project Ideas Portal</h1>
      <p>Register or login, then submit project ideas.</p>

      {error && <p className="error">{error}</p>}

      <section className="card">
        {!token ? (
          <>
            <h2>Account</h2>
            <div className="row">
              <input placeholder="Email" value={authEmail} onChange={(e) => setAuthEmail(e.target.value)} />
              <input placeholder="Password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
            </div>
            <div className="row" style={{ marginTop: 10 }}>
              <button className="btn-dark" onClick={() => handleAuth('register')}>Register</button>
              <button className="btn-mid" onClick={() => handleAuth('login')}>Login</button>
            </div>
          </>
        ) : (
          <>
            <div className="row" style={{ justifyContent: 'space-between', alignItems: 'center' }}>
              <h2>Submit idea</h2>
              <button className="btn-outline" onClick={logout}>Logout ({email})</button>
            </div>
            <div className="col">
              <input placeholder="Idea title" value={title} onChange={(e) => setTitle(e.target.value)} />
              <textarea placeholder="Idea description" value={description} onChange={(e) => setDescription(e.target.value)} />
              <button className="btn-blue" onClick={submitIdea}>Create idea</button>
            </div>
          </>
        )}
      </section>

      <section className="card">
        <h2>Ideas</h2>
        <ul className="idea-list">
          {ideas.map((idea) => (
            <li key={idea.id} className="idea-item">
              <div><strong>{idea.title}</strong></div>
              <p>{idea.description}</p>
              <p className="meta">by {idea.userEmail} · {new Date(idea.createdAt).toLocaleString()}</p>
            </li>
          ))}
        </ul>
      </section>
    </main>
  )
}

createRoot(document.getElementById('root')!).render(<App />)
