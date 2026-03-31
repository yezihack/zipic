<template>
  <div class="image-compare" ref="containerRef" @mousemove="handleMouseMove" @touchmove="handleTouchMove">
    <!-- After (Compressed) - right side, full width -->
    <div class="image-after">
      <img :src="afterSrc" :alt="afterLabel" @load="onImageLoad" draggable="false" />
      <span class="label-after">{{ afterLabel }}</span>
    </div>

    <!-- Before (Original) - left side, clipped -->
    <div class="image-before" :style="{ width: sliderPosition + '%' }">
      <img :src="beforeSrc" :alt="beforeLabel" draggable="false" />
      <span class="label-before">{{ beforeLabel }}</span>
    </div>

    <!-- Slider handle -->
    <div
      class="slider-handle"
      :style="{ left: sliderPosition + '%' }"
      @mousedown="startDrag"
      @touchstart="startDrag"
    >
      <div class="slider-line"></div>
      <div class="slider-button">
        <span class="slider-arrows">&lt;&gt;</span>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = withDefaults(
  defineProps<{
    beforeSrc: string
    afterSrc: string
    beforeLabel?: string
    afterLabel?: string
    initialPosition?: number
  }>(),
  {
    beforeLabel: 'Original',
    afterLabel: 'Compressed',
    initialPosition: 50
  }
)

const containerRef = ref<HTMLDivElement | null>(null)
const sliderPosition = ref(props.initialPosition)
const isDragging = ref(false)
const imageLoaded = ref(false)

function onImageLoad(): void {
  imageLoaded.value = true
}

function startDrag(event: MouseEvent | TouchEvent): void {
  event.preventDefault()
  isDragging.value = true
}

function stopDrag(): void {
  isDragging.value = false
}

function handleMouseMove(event: MouseEvent): void {
  if (!isDragging.value || !containerRef.value) return
  updatePosition(event.clientX)
}

function handleTouchMove(event: TouchEvent): void {
  if (!isDragging.value || !containerRef.value) return
  if (event.touches.length > 0) {
    updatePosition(event.touches[0].clientX)
  }
}

function updatePosition(clientX: number): void {
  const container = containerRef.value
  if (!container) return

  const rect = container.getBoundingClientRect()
  const x = clientX - rect.left
  const percentage = Math.max(0, Math.min(100, (x / rect.width) * 100))
  sliderPosition.value = percentage
}

function handleGlobalMouseMove(event: MouseEvent): void {
  if (!isDragging.value || !containerRef.value) return
  updatePosition(event.clientX)
}

function handleGlobalTouchMove(event: TouchEvent): void {
  if (!isDragging.value || !containerRef.value) return
  if (event.touches.length > 0) {
    updatePosition(event.touches[0].clientX)
  }
}

function handleGlobalMouseUp(): void {
  stopDrag()
}

onMounted(() => {
  document.addEventListener('mousemove', handleGlobalMouseMove)
  document.addEventListener('mouseup', handleGlobalMouseUp)
  document.addEventListener('touchmove', handleGlobalTouchMove)
  document.addEventListener('touchend', handleGlobalMouseUp)
})

onUnmounted(() => {
  document.removeEventListener('mousemove', handleGlobalMouseMove)
  document.removeEventListener('mouseup', handleGlobalMouseUp)
  document.removeEventListener('touchmove', handleGlobalTouchMove)
  document.removeEventListener('touchend', handleGlobalMouseUp)
})
</script>

<style scoped>
.image-compare {
  position: relative;
  width: 100%;
  overflow: hidden;
  border-radius: 8px;
  cursor: ew-resize;
  user-select: none;
  background: #f3f4f6;
}

/* Adaptive height based on container, not fixed aspect ratio */
.image-compare img {
  display: block;
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
}

.image-after {
  position: absolute;
  top: 0;
  left: 0;
  width: 100%;
  height: 100%;
}

.image-before {
  position: absolute;
  top: 0;
  left: 0;
  height: 100%;
  overflow: hidden;
  z-index: 1;
}

.image-before img {
  position: absolute;
  top: 0;
  left: 0;
  width: auto;
  min-width: 100%;
  height: 100%;
  object-fit: contain;
}

.label-before,
.label-after {
  position: absolute;
  bottom: 12px;
  padding: 4px 12px;
  background: rgba(0, 0, 0, 0.6);
  color: white;
  font-size: 12px;
  font-weight: 600;
  border-radius: 4px;
  pointer-events: none;
  white-space: nowrap;
}

.label-before {
  left: 12px;
}

.label-after {
  right: 12px;
}

.slider-handle {
  position: absolute;
  top: 0;
  bottom: 0;
  width: 4px;
  z-index: 10;
  transform: translateX(-50%);
  cursor: ew-resize;
}

.slider-line {
  position: absolute;
  top: 0;
  bottom: 0;
  left: 50%;
  width: 2px;
  background: white;
  transform: translateX(-50%);
  box-shadow: 0 0 4px rgba(0, 0, 0, 0.3);
}

.slider-button {
  position: absolute;
  top: 50%;
  left: 50%;
  transform: translate(-50%, -50%);
  width: 40px;
  height: 40px;
  background: white;
  border-radius: 50%;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.3);
  color: #374151;
  font-weight: bold;
}

.slider-arrows {
  font-size: 14px;
  letter-spacing: 2px;
}
</style>