(function () {
    // Configuration
    const config = {
        customerID: "{{.CustomerID}}",
        wsEndpoint: "{{.WSEndpoint}}",
        brandColors: {{.BrandColors}},
        logoURL: "{{.LogoURL}}",
        modelPath: "{{.ModelPath}}",
        animations: {{.Animations}}
    };

// Create chat widget container
const widgetContainer = document.getElementById('chat-widget');
if (!widgetContainer) {
    console.error('Chat widget container not found');
    return;
}

// Inject styles
const style = document.createElement('style');
const colors = JSON.parse(config.brandColors || '{}');
style.textContent = `
        .chat-widget {
            position: fixed;
            bottom: 20px;
            right: 20px;
            width: 350px;
            height: 500px;
            background: ${colors.background || '#ffffff'};
            border-radius: 10px;
            box-shadow: 0 0 20px rgba(0,0,0,0.1);
            display: flex;
            flex-direction: column;
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            z-index: 9999;
        }
        .chat-header {
            background: ${colors.primary || '#007bff'};
            color: ${colors.primaryText || '#ffffff'};
            padding: 15px;
            border-radius: 10px 10px 0 0;
            display: flex;
            align-items: center;
            gap: 10px;
        }
        .chat-logo {
            width: 30px;
            height: 30px;
            border-radius: 50%;
        }
        .chat-messages {
            flex: 1;
            overflow-y: auto;
            padding: 15px;
        }
        .chat-message {
            margin-bottom: 10px;
            padding: 10px;
            border-radius: 8px;
            max-width: 80%;
        }
        .chat-message.user {
            background: ${colors.userMessage || '#e3f2fd'};
            margin-left: auto;
        }
        .chat-message.assistant {
            background: ${colors.assistantMessage || '#f5f5f5'};
        }
        .chat-input-container {
            padding: 15px;
            border-top: 1px solid #eee;
        }
        .chat-input {
            width: 100%;
            padding: 10px;
            border: 1px solid #ddd;
            border-radius: 5px;
            font-size: 14px;
        }
        .chat-3d-container {
            height: 150px;
            background: #f0f0f0;
            position: relative;
            display: ${config.modelPath ? 'block' : 'none'};
        }
    `;
document.head.appendChild(style);

// Create chat HTML
widgetContainer.innerHTML = `
        <div class="chat-widget">
            <div class="chat-header">
                ${config.logoURL ? '<img src="' + config.logoURL + '" class="chat-logo">' : ''}
                <h3>Chat Assistant</h3>
            </div>
            ${config.modelPath ? '<div class="chat-3d-container" id="three-container"></div>' : ''}
            <div class="chat-messages" id="chat-messages">
                <div class="chat-message assistant">
                    Hello! How can I help you today?
                </div>
            </div>
            <div class="chat-input-container">
                <input type="text" class="chat-input" id="chat-input" placeholder="Type your message...">
            </div>
        </div>
    `;

// Initialize Three.js if model is available
if (config.modelPath && typeof THREE !== 'undefined') {
    initializeThreeJS();
}

function initializeThreeJS() {
    const threeContainer = document.getElementById('three-container');
    const scene = new THREE.Scene();
    const camera = new THREE.PerspectiveCamera(75, threeContainer.clientWidth / threeContainer.clientHeight, 0.1, 1000);
    const renderer = new THREE.WebGLRenderer({ alpha: true });
    renderer.setSize(threeContainer.clientWidth, threeContainer.clientHeight);
    threeContainer.appendChild(renderer.domElement);

    // Load 3D model
    let model, mixer;
    const loader = new THREE.GLTFLoader();
    if (config.modelPath) {
        loader.load(config.modelPath, (gltf) => {
            model = gltf.scene;
            scene.add(model);

            // Setup animations
            if (gltf.animations && gltf.animations.length) {
                mixer = new THREE.AnimationMixer(model);
                const animations = JSON.parse(config.animations || '{}');

                // Store animation clips
                window.chatAnimations = {};
                gltf.animations.forEach(clip => {
                    window.chatAnimations[ clip.name ] = mixer.clipAction(clip);
                });
            }
        });
    }

    // Setup camera and lights
    camera.position.z = 5;
    const light = new THREE.DirectionalLight(0xffffff, 1);
    light.position.set(0, 1, 1);
    scene.add(light);
    scene.add(new THREE.AmbientLight(0x404040));

    // Animation loop
    const clock = new THREE.Clock();
    function animate() {
        requestAnimationFrame(animate);
        if (mixer) mixer.update(clock.getDelta());
        renderer.render(scene, camera);
    }
    animate();
}

// WebSocket connection
const ws = new WebSocket(config.wsEndpoint);
const messagesContainer = document.getElementById('chat-messages');
const chatInput = document.getElementById('chat-input');

// Generate or retrieve user ID
let userID = localStorage.getItem('chat-user-id');
if (!userID) {
    userID = 'user-' + Date.now() + '-' + Math.random().toString(36).substr(2, 9);
    localStorage.setItem('chat-user-id', userID);
}

ws.onopen = () => {
    console.log('Connected to chat server');
};

ws.onmessage = (event) => {
    const data = JSON.parse(event.data);

    // Add message to chat
    const messageEl = document.createElement('div');
    messageEl.className = 'chat-message assistant';
    messageEl.textContent = data.content;
    messagesContainer.appendChild(messageEl);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;

    // Play animation based on emotion
    if (data.emotion && window.chatAnimations) {
        const animations = JSON.parse(config.animations || '{}');
        const animConfig = animations[ data.emotion ];
        if (animConfig && window.chatAnimations[ animConfig.name ]) {
            const action = window.chatAnimations[ animConfig.name ];
            action.reset();
            action.play();
        }
    }
};

ws.onerror = (error) => {
    console.error('WebSocket error:', error);
};

ws.onclose = () => {
    console.log('WebSocket connection closed');
};

// Handle input
chatInput.addEventListener('keypress', (e) => {
    if (e.key === 'Enter' && chatInput.value.trim()) {
        sendMessage();
    }
});

function sendMessage() {
    const message = chatInput.value.trim();
    if (!message) return;

    // Add user message to chat
    const messageEl = document.createElement('div');
    messageEl.className = 'chat-message user';
    messageEl.textContent = message;
    messagesContainer.appendChild(messageEl);
    messagesContainer.scrollTop = messagesContainer.scrollHeight;

    // Send message via WebSocket
    if (ws.readyState === WebSocket.OPEN) {
        ws.send(JSON.stringify({
            content: message,
            userID: userID
        }));
    }

    chatInput.value = '';
}
}) ();
