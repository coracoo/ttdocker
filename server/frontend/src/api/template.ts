import axios from 'axios'

const api = axios.create({
  baseURL: '/api'
})

export const templateApi = {
  list() {
    return api.get('/templates')
  },

  get(id: string) {
    return api.get(`/templates/${id}`)
  },

  create(data: any) {
    return api.post('/templates', data)
  },

  update(id: string, data: any) {
    return api.put(`/templates/${id}`, data)
  },

  delete(id: string) {
    return api.delete(`/templates/${id}`)
  }
}