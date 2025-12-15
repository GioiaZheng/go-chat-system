// axios.js configures the shared Axios instance used by the frontend API helpers.
import axios from 'axios';

const instance = axios.create({
  baseURL: __API_URL__,
  timeout: 1000 * 5,
});

export default instance;