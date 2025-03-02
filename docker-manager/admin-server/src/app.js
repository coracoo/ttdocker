const express = require('express');
const cors = require('cors');
const bodyParser = require('body-parser');
const fs = require('fs');
const path = require('path');
const multer = require('multer');
const yaml = require('js-yaml');

const app = express();
const PORT = process.env.PORT || 3001;

// 中间件
app.use(cors());
app.use(bodyParser.json());
app.use(express.static(path.join(__dirname, 'public')));

// 配置文件上传
const storage = multer.diskStorage({
  destination: (req, file, cb) => {
    cb(null, path.join(__dirname, 'public', 'uploads'));
  },
  filename: (req, file, cb) => {
    cb(null, Date.now() + '-' + file.originalname);
  }
});
const upload = multer({ storage });

// 应用数据存储路径
const APPS_DIR = path.join(__dirname, 'data', 'apps');

// 确保目录存在
if (!fs.existsSync(APPS_DIR)) {
  fs.mkdirSync(APPS_DIR, { recursive: true });
}
if (!fs.existsSync(path.join(__dirname, 'public', 'uploads'))) {
  fs.mkdirSync(path.join(__dirname, 'public', 'uploads'), { recursive: true });
}

// 获取所有应用
app.get('/api/apps', (req, res) => {
  try {
    const files = fs.readdirSync(APPS_DIR);
    const apps = files
      .filter(file => file.endsWith('.json'))
      .map(file => {
        const content = fs.readFileSync(path.join(APPS_DIR, file), 'utf8');
        return JSON.parse(content);
      });
    res.json(apps);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 获取单个应用
app.get('/api/apps/:id', (req, res) => {
  try {
    const appPath = path.join(APPS_DIR, `${req.params.id}.json`);
    if (!fs.existsSync(appPath)) {
      return res.status(404).json({ error: '应用不存在' });
    }
    const content = fs.readFileSync(appPath, 'utf8');
    res.json(JSON.parse(content));
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 创建/更新应用
app.post('/api/apps', (req, res) => {
  try {
    const app = req.body;
    if (!app.id) {
      return res.status(400).json({ error: '应用ID是必需的' });
    }
    
    fs.writeFileSync(
      path.join(APPS_DIR, `${app.id}.json`),
      JSON.stringify(app, null, 2),
      'utf8'
    );
    
    res.json({ success: true, message: '应用已保存' });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 删除应用
app.delete('/api/apps/:id', (req, res) => {
  try {
    const appPath = path.join(APPS_DIR, `${req.params.id}.json`);
    if (!fs.existsSync(appPath)) {
      return res.status(404).json({ error: '应用不存在' });
    }
    
    fs.unlinkSync(appPath);
    res.json({ success: true, message: '应用已删除' });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 上传Logo
app.post('/api/upload/logo', upload.single('logo'), (req, res) => {
  try {
    if (!req.file) {
      return res.status(400).json({ error: '没有上传文件' });
    }
    
    const fileUrl = `/uploads/${req.file.filename}`;
    res.json({ url: fileUrl });
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

// 从YAML转换为JSON
app.post('/api/convert', upload.single('yaml'), (req, res) => {
  try {
    if (!req.file) {
      return res.status(400).json({ error: '没有上传文件' });
    }
    
    const yamlContent = fs.readFileSync(req.file.path, 'utf8');
    const composeConfig = yaml.load(yamlContent);
    
    // 提取第一个服务作为主服务
    const serviceName = Object.keys(composeConfig.services)[0];
    const service = composeConfig.services[serviceName];
    
    // 构建应用JSON
    const app = {
      id: serviceName,
      name: serviceName,
      description: `${serviceName} 应用`,
      category: "其他",
      version: "1.0.0",
      logo: "",
      author: "Docker Manager",
      website: "",
      tags: [],
      ports: [],
      volumes: [],
      environment: [],
      compose: composeConfig
    };
    
    // 处理端口
    if (service.ports) {
      app.ports = service.ports.map(port => {
        const [host, container] = port.split(':');
        return {
          container: parseInt(container),
          host: parseInt(host),
          description: "端口映射"
        };
      });
    }
    
    // 处理卷
    if (service.volumes) {
      app.volumes = service.volumes.map(volume => {
        const [host, container] = volume.split(':');
        return {
          container,
          host,
          description: "数据卷"
        };
      });
    }
    
    // 处理环境变量
    if (service.environment) {
      app.environment = service.environment.map(env => {
        const [name, value] = env.split('=');
        return {
          name,
          value,
          description: "环境变量"
        };
      });
    }
    
    res.json(app);
  } catch (error) {
    res.status(500).json({ error: error.message });
  }
});

app.listen(PORT, () => {
  console.log(`Admin server running on port ${PORT}`);
});