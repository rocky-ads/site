// image-preview.js: Clean image upload system using DOM as single source of truth
let draggedElement = null;
let dragAndDropSetup = false;
let touchStartX = 0;
let touchStartY = 0;
let isDragging = false;
let pendingAdditions = 0; // Track pending async additions to prevent race conditions

function existingImageCount() {
	return typeof EXISTING_IMAGE_COUNT !== 'undefined' ? EXISTING_IMAGE_COUNT : 0;
}

function imageAppendMode() {
	return typeof IMAGE_APPEND_MODE !== 'undefined' && IMAGE_APPEND_MODE;
}

function totalImageCount() {
	const preview = document.getElementById('image-preview');
	const newCount = preview ? preview.children.length : 0;
	return existingImageCount() + newCount + pendingAdditions;
}

function canAddMoreImages() {
	return totalImageCount() < MAX_IMAGES_PER_AD;
}

function handleUploadClick() {
	if (canAddMoreImages()) {
		document.getElementById('images').click();
	}
}

function previewImages(input) {
	const preview = document.getElementById('image-preview');
	
	// Add new files to the DOM
	if (input.files && input.files.length > 0) {
		let skippedCount = 0;
		Array.from(input.files).forEach(file => {
			if (totalImageCount() >= MAX_IMAGES_PER_AD) {
				skippedCount++;
				return;
			}
			
			// Check if file is already in our collection (by name and size)
			const isDuplicate = Array.from(preview.children).some(thumbnail => 
				thumbnail.fileReference.name === file.name && thumbnail.fileReference.size === file.size
			);
			if (!isDuplicate) {
				pendingAdditions++;
				addThumbnailToDOM(preview, file);
			}
		});
		
		if (skippedCount > 0) {
			alert(`Maximum of ${MAX_IMAGES_PER_AD} images allowed per ad. ${skippedCount} image(s) were not added.`);
		}
	}
	
	if (!dragAndDropSetup && !imageAppendMode()) {
		setupDragAndDrop(preview);
		dragAndDropSetup = true;
	}
	
	updateUploadAreaText();
	updateFileInputFromDOM();
	toggleUploadContent();
}

function addThumbnailToDOM(preview, file) {
	const container = document.createElement('div');
	container.className = imageAppendMode()
		? 'inline-block'
		: 'inline-block cursor-move';
	container.style.margin = '2px';
	container.style.padding = '3px';
	container.style.borderRadius = '8px';
	if (!imageAppendMode()) {
		container.setAttribute('draggable', 'true');
	}
	container.fileReference = file;
	
	const thumbnail = document.createElement('div');
	thumbnail.className = 'relative';
	thumbnail.style.width = '90px';
	thumbnail.style.height = '90px';
	
	container.onclick = function(event) {
		event.stopPropagation();
	};
	
	const reader = new FileReader();
	reader.onload = function(e) {
		const img = document.createElement('img');
		img.className = 'object-cover rounded';
		img.style.display = 'block';
		img.style.width = '90px';
		img.style.height = '90px';
		img.style.borderRadius = '6px';
		img.src = e.target.result;
		img.alt = file.name;
		
		thumbnail.appendChild(img);
		if (!imageAppendMode()) {
			const deleteBtn = document.createElement('button');
			deleteBtn.className = 'absolute top-0 right-0 w-6 h-6 bg-red-500 text-white rounded-full text-xs hover:bg-red-600';
			deleteBtn.innerHTML = '×';
			deleteBtn.onclick = function(event) {
				event.stopPropagation();
				container.remove();
				updateUploadAreaText();
				updateFileInputFromDOM();
				toggleUploadContent();
			};
			thumbnail.appendChild(deleteBtn);
		}
		container.appendChild(thumbnail);
		preview.appendChild(container);
		
		pendingAdditions--;
		
		updateUploadAreaText();
		updateFileInputFromDOM();
		toggleUploadContent();
	};
	reader.readAsDataURL(file);
}

function updateFileInputFromDOM() {
	const preview = document.getElementById('image-preview');
	const input = document.getElementById('images');
	const files = Array.from(preview.children).map(thumbnail => thumbnail.fileReference);
	
	const dt = new DataTransfer();
	files.forEach(file => {
		dt.items.add(file);
	});
	input.files = dt.files;
}

function setupDragAndDrop(preview) {
	preview.addEventListener('dragstart', function(e) {
		if (e.target.closest('[draggable="true"]')) {
			draggedElement = e.target.closest('[draggable="true"]');
			draggedElement.style.opacity = '0.5';
			e.dataTransfer.effectAllowed = 'move';
		}
	});
	
	preview.addEventListener('dragover', function(e) {
		e.preventDefault();
		const over = e.target.closest('[draggable="true"]');
		if (draggedElement && over && draggedElement !== over) {
			over.style.backgroundColor = '#22c55e';
		}
	});
	
	preview.addEventListener('dragleave', function(e) {
		const over = e.target.closest('[draggable="true"]');
		if (over) {
			over.style.backgroundColor = '';
		}
	});
	
	preview.addEventListener('drop', function(e) {
		e.preventDefault();
		const over = e.target.closest('[draggable="true"]');
		if (draggedElement && over && draggedElement !== over) {
			over.style.backgroundColor = '';
			
			const allThumbnails = Array.from(preview.children);
			const draggedIndex = allThumbnails.indexOf(draggedElement);
			const overIndex = allThumbnails.indexOf(over);
			
			if (draggedIndex < overIndex) {
				preview.insertBefore(draggedElement, over.nextSibling);
			} else {
				preview.insertBefore(draggedElement, over);
			}
			
			updateFileInputFromDOM();
		}
	});
	
	preview.addEventListener('dragend', function() {
		Array.from(preview.children).forEach(el => {
			el.style.backgroundColor = '';
			el.style.opacity = '';
		});
		draggedElement = null;
	});
	
	preview.addEventListener('touchstart', function(e) {
		const thumbnail = e.target.closest('[draggable="true"]');
		if (thumbnail) {
			draggedElement = thumbnail;
			const touch = e.touches[0];
			touchStartX = touch.clientX;
			touchStartY = touch.clientY;
			isDragging = false;
			
			setTimeout(() => {
				if (draggedElement) {
					draggedElement.style.opacity = '0.5';
				}
			}, 100);
		}
	}, { passive: false });
	
	preview.addEventListener('touchmove', function(e) {
		if (!draggedElement) return;
		
		const touch = e.touches[0];
		const deltaX = Math.abs(touch.clientX - touchStartX);
		const deltaY = Math.abs(touch.clientY - touchStartY);
		
		if (deltaX > 10 || deltaY > 10) {
			isDragging = true;
			e.preventDefault();
		}
		
		if (isDragging) {
			const elementAtPoint = document.elementFromPoint(touch.clientX, touch.clientY);
			const over = elementAtPoint?.closest('[draggable="true"]');
			
			Array.from(preview.children).forEach(el => {
				if (el !== draggedElement) {
					el.style.backgroundColor = '';
				}
			});
			
			if (over && over !== draggedElement) {
				over.style.backgroundColor = '#22c55e';
			}
		}
	}, { passive: false });
	
	preview.addEventListener('touchend', function(e) {
		if (!draggedElement) return;
		
		if (isDragging) {
			const touch = e.changedTouches[0];
			const elementAtPoint = document.elementFromPoint(touch.clientX, touch.clientY);
			const over = elementAtPoint?.closest('[draggable="true"]');
			
			if (over && over !== draggedElement) {
				const allThumbnails = Array.from(preview.children);
				const draggedIndex = allThumbnails.indexOf(draggedElement);
				const overIndex = allThumbnails.indexOf(over);
				
				if (draggedIndex < overIndex) {
					preview.insertBefore(draggedElement, over.nextSibling);
				} else {
					preview.insertBefore(draggedElement, over);
				}
				
				updateFileInputFromDOM();
			}
		}
		
		Array.from(preview.children).forEach(el => {
			el.style.backgroundColor = '';
			el.style.opacity = '';
		});
		draggedElement = null;
		isDragging = false;
	}, { passive: false });
}

function handleDrop(event) {
	const files = Array.from(event.dataTransfer.files).filter(file => 
		file.type.startsWith('image/')
	);
	
	if (files.length > 0) {
		const preview = document.getElementById('image-preview');
		let skippedCount = 0;
		files.forEach(file => {
			if (totalImageCount() >= MAX_IMAGES_PER_AD) {
				skippedCount++;
				return;
			}
			
			const isDuplicate = Array.from(preview.children).some(thumbnail => 
				thumbnail.fileReference.name === file.name && thumbnail.fileReference.size === file.size
			);
			if (!isDuplicate) {
				pendingAdditions++;
				addThumbnailToDOM(preview, file);
			}
		});
		
		if (skippedCount > 0) {
			alert(`Maximum of ${MAX_IMAGES_PER_AD} images allowed per ad. ${skippedCount} image(s) were not added.`);
		}
	}
}

function toggleUploadContent() {
	const uploadContent = document.getElementById('upload-content');
	const imagePreview = document.getElementById('image-preview');
	const preview = document.getElementById('image-preview');
	const fileCount = preview.children.length;
	
	if (fileCount === 0) {
		uploadContent.classList.remove('hidden');
		imagePreview.classList.add('hidden');
	} else {
		uploadContent.classList.remove('hidden');
		imagePreview.classList.remove('hidden');
	}
}

function updateUploadAreaText() {
	const uploadContent = document.getElementById('upload-content');
	const uploadArea = document.getElementById('upload-area');
	if (!uploadContent) {
		return;
	}
	const titleElement = uploadContent.querySelector('.text-lg');
	const subtitleElement = uploadContent.querySelector('.text-sm');
	
	const preview = document.getElementById('image-preview');
	const newCount = preview.children.length;
	const totalCount = existingImageCount() + newCount;
	const remaining = MAX_IMAGES_PER_AD - totalCount;
	
	if (imageAppendMode()) {
		if (totalCount >= MAX_IMAGES_PER_AD) {
			titleElement.textContent = `Maximum ${MAX_IMAGES_PER_AD} Images Reached`;
			subtitleElement.textContent = 'No more images can be added';
			uploadContent.classList.add('opacity-50', 'cursor-not-allowed');
			uploadArea.style.cursor = 'not-allowed';
		} else if (newCount === 0) {
			titleElement.textContent = 'Add More Images';
			subtitleElement.textContent =
				`${totalCount} of ${MAX_IMAGES_PER_AD} (${remaining} remaining)`;
			uploadContent.classList.remove('opacity-50', 'cursor-not-allowed');
			uploadArea.style.cursor = 'pointer';
		} else {
			titleElement.textContent = 'Add More Images';
			subtitleElement.textContent =
				`${totalCount} of ${MAX_IMAGES_PER_AD} (${remaining} remaining)`;
			uploadContent.classList.remove('opacity-50', 'cursor-not-allowed');
			uploadArea.style.cursor = 'pointer';
		}
		return;
	}

	if (newCount === 0) {
		titleElement.textContent = 'Upload Images';
		subtitleElement.textContent =
			`Click to browse or drag and drop (up to ${MAX_IMAGES_PER_AD} images)`;
		uploadContent.classList.remove('opacity-50', 'cursor-not-allowed');
		uploadArea.style.cursor = 'pointer';
	} else if (newCount >= MAX_IMAGES_PER_AD) {
		titleElement.textContent = `Maximum ${MAX_IMAGES_PER_AD} Images Reached`;
		subtitleElement.textContent = 'Delete an image to add more';
		uploadContent.classList.add('opacity-50', 'cursor-not-allowed');
		uploadArea.style.cursor = 'not-allowed';
	} else {
		titleElement.textContent = 'Add More Images';
		subtitleElement.textContent =
			`${newCount} of ${MAX_IMAGES_PER_AD} (${remaining} remaining)`;
		uploadContent.classList.remove('opacity-50', 'cursor-not-allowed');
		uploadArea.style.cursor = 'pointer';
	}
}
