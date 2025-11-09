class ImageProcessor {
    constructor() {
        this.baseUrl = window.location.origin;
        this.initializeEventListeners();
    }

    initializeEventListeners() {
        const uploadForm = document.getElementById('uploadForm');
        if (uploadForm) {
            uploadForm.addEventListener('submit', (e) => {
                e.preventDefault();
                this.uploadImage();
            });
        }
    }

    async uploadImage() {
        const formData = new FormData(document.getElementById('uploadForm'));
        const uploadBtn = document.getElementById('uploadBtn');

        try {
            uploadBtn.disabled = true;
            uploadBtn.textContent = 'Загрузка...';

            const response = await fetch(`${this.baseUrl}/upload`, {
                method: 'POST',
                body: formData
            });

            if (response.ok) {
                const result = await response.json();
                this.addImageToList(result);
                document.getElementById('uploadForm').reset();
                alert('Изображение загружено и обрабатывается!');
            } else {
                const errorData = await response.json().catch(() => ({ error: 'Ошибка загрузки' }));
                throw new Error(errorData.details || errorData.error);
            }
        } catch (error) {
            alert(`Ошибка: ${error.message}`);
        } finally {
            uploadBtn.disabled = false;
            uploadBtn.textContent = 'Загрузить и обработать';
        }
    }

    addImageToList(imageData) {
        const container = document.getElementById('imagesList');
        const emptyState = document.getElementById('emptyState');

        // Скрываем пустое состояние
        if (emptyState) {
            emptyState.style.display = 'none';
        }

        // Создаем элемент для изображения
        const imageElement = document.createElement('div');
        imageElement.className = 'image-item';
        imageElement.id = `image-${imageData.image_id}`;

        // Добавляем информацию об опциях обработки, если они есть
        let optionsInfo = '';
        if (imageData.processing_options) {
            const opts = imageData.processing_options;
            optionsInfo = `
                <div><strong>Опции:</strong> 
                    ${opts.width ? `Ш:${opts.width}px ` : ''}
                    ${opts.height ? `В:${opts.height}px ` : ''}
                    ${opts.quality ? `Кач:${opts.quality} ` : ''}
                    ${opts.format ? `Формат:${opts.format} ` : ''}
                    ${opts.watermark_text ? `Вод.знак:${opts.watermark_text} ` : ''}
                    ${opts.thumbnail ? 'Миниатюра' : ''}
                </div>
            `;
        }

        imageElement.innerHTML = `
            <div class="image-info">
                <strong>ID:</strong> ${imageData.image_id}<br>
                <strong>Статус:</strong> <span class="status" id="status-${imageData.image_id}">⏳ Обрабатывается...</span><br>
                <strong>Ссылка:</strong> <span id="link-${imageData.image_id}">-</span>
                ${optionsInfo}
            </div>
            <div class="image-actions">
                <button onclick="checkStatus('${imageData.image_id}')">Проверить статус</button>
                <button onclick="deleteImage('${imageData.image_id}')" class="btn-delete">Удалить</button>
            </div>
        `;

        if (container) {
            container.appendChild(imageElement);
        }

        // Запускаем проверку статуса
        this.pollImageStatus(imageData.image_id);
    }

    async pollImageStatus(imageId) {
        const checkStatus = async () => {
            try {
                const response = await fetch(`${this.baseUrl}/image/${imageId}`);

                if (response.ok) {
                    const contentType = response.headers.get('content-type');

                    if (contentType && contentType.includes('application/json')) {
                        // JSON ответ - изображение еще обрабатывается
                        const data = await response.json();
                        this.updateImageStatus(imageId, data.status, data.image_url);

                        if (data.status === 'processing' || data.status === 'uploaded') {
                            // Продолжаем опрос
                            setTimeout(() => checkStatus(), 2000);
                        }
                    } else if (contentType && contentType.includes('image/')) {
                        // Бинарные данные - готово!
                        this.updateImageStatus(imageId, 'completed', `${this.baseUrl}/image/${imageId}`);
                    }
                }
            } catch (error) {
                console.error('Status check error:', error);
                // Продолжаем опрос даже при ошибке
                setTimeout(() => checkStatus(), 2000);
            }
        };

        checkStatus();
    }

    updateImageStatus(imageId, status, imageUrl = null) {
        const statusElement = document.getElementById(`status-${imageId}`);
        const linkElement = document.getElementById(`link-${imageId}`);

        if (statusElement) {
            const statusText = {
                'processing': '⏳ Обрабатывается...',
                'completed': '✅ Готово',
                'failed': '❌ Ошибка',
                'uploaded': '📤 Загружено'
            };

            statusElement.textContent = statusText[status] || status;
            statusElement.className = `status status-${status}`;
        }

        if (linkElement && imageUrl) {
            linkElement.innerHTML = `<a href="${imageUrl}" target="_blank" download="image-${imageId}">Скачать изображение</a>`;
        }
    }

    async deleteImage(imageId) {
        if (!confirm('Удалить это изображение?')) return;

        try {
            const response = await fetch(`${this.baseUrl}/image/${imageId}`, {
                method: 'DELETE'
            });

            if (response.ok) {
                // Удаляем элемент из DOM
                const imageElement = document.getElementById(`image-${imageId}`);
                if (imageElement) {
                    imageElement.remove();
                }

                // Показываем пустое состояние, если изображений не осталось
                const container = document.getElementById('imagesList');
                const emptyState = document.getElementById('emptyState');
                if (container && container.children.length === 0 && emptyState) {
                    emptyState.style.display = 'block';
                }

                alert('Изображение удалено успешно!');
            } else {
                const errorData = await response.json().catch(() => ({ error: 'Ошибка удаления' }));
                throw new Error(errorData.details || errorData.error);
            }
        } catch (error) {
            alert(`Ошибка при удалении: ${error.message}`);
        }
    }
}

// Глобальная функция для ручной проверки статуса
function checkStatus(imageId) {
    fetch(`${window.app.baseUrl}/image/${imageId}`)
        .then(response => {
            if (response.ok) {
                const contentType = response.headers.get('content-type');

                if (contentType && contentType.includes('application/json')) {
                    return response.json().then(data => {
                        window.app.updateImageStatus(imageId, data.status, data.image_url);
                    });
                } else if (contentType && contentType.includes('image/')) {
                    window.app.updateImageStatus(imageId, 'completed', `${window.app.baseUrl}/image/${imageId}`);
                }
            }
        })
        .catch(error => {
            console.error('Manual status check error:', error);
        });
}

// Глобальная функция для удаления изображения
function deleteImage(imageId) {
    if (window.app && window.app.deleteImage) {
        window.app.deleteImage(imageId);
    }
}

// Функция для переключения видимости опций
function toggleOptions() {
    const options = document.getElementById('processingOptions');
    const button = document.querySelector('.toggle-options');

    if (options.style.display === 'none') {
        options.style.display = 'block';
        button.textContent = '△ Скрыть опции обработки';
    } else {
        options.style.display = 'none';
        button.textContent = '▽ Показать опции обработки';
    }
}

// Инициализация приложения
const app = new ImageProcessor();
window.app = app;