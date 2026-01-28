// Package chromebrowser provides deterministic browser scripts.
package chromebrowser

import "fmt"

// DisableAnimationsCSS contains CSS to disable all animations and transitions.
const DisableAnimationsCSS = `
*, *::before, *::after {
  animation: none !important;
  animation-duration: 0s !important;
  animation-delay: 0s !important;
  transition: none !important;
  transition-duration: 0s !important;
  transition-delay: 0s !important;
  caret-color: transparent !important;
}
html {
  scroll-behavior: auto !important;
}
@media (prefers-reduced-motion: no-preference) {
  :root {
    --force-no-motion: 1;
  }
}
`

// GenerateMockTimeScript generates a script to fix Date, Math.random, and Performance.now.
func GenerateMockTimeScript(fixedTime string) string {
	return fmt.Sprintf(`
(() => {
  // Fix Date
  const fixedTimestamp = new Date('%s').valueOf();
  const OriginalDate = Date;

  window.Date = class extends OriginalDate {
    constructor(...args) {
      if (args.length === 0) {
        super(fixedTimestamp);
      } else {
        super(...args);
      }
    }

    static now() {
      return fixedTimestamp;
    }
  };

  // Copy static methods
  Object.setPrototypeOf(window.Date, OriginalDate);
  Object.setPrototypeOf(window.Date.prototype, OriginalDate.prototype);

  // Fix Math.random to always return 0.5 for true determinism
  Math.random = function() {
    return 0.5;
  };

  // Fix Performance.now
  if (window.performance && window.performance.now) {
    let offset = 0;
    performance.now = function() {
      return offset;
    };
  }
})();
`, fixedTime)
}

// DisableAutoplayScript disables autoplay features and freezes video/audio elements.
const DisableAutoplayScript = `
(() => {
  // Global flag for applications to check
  window.__E2E_DISABLE_AUTOPLAY__ = true;

  // Helper function to freeze a media element
  function freezeMedia(media) {
    try {
      media.pause();
      media.currentTime = 0;
      media.autoplay = false;
      media.loop = false;
      media.removeAttribute('autoplay');
    } catch (e) {}
  }

  // Disable video/audio autoplay and freeze at time 0
  HTMLMediaElement.prototype.play = function() {
    this.pause();
    this.currentTime = 0;
    return Promise.resolve();
  };

  // Prevent autoplay attribute from being set
  Object.defineProperty(HTMLMediaElement.prototype, 'autoplay', {
    get() { return false; },
    set() {},
    configurable: true
  });

  // Prevent loop attribute
  Object.defineProperty(HTMLMediaElement.prototype, 'loop', {
    get() { return false; },
    set() {},
    configurable: true
  });

  // Override load method to pause immediately
  const originalLoad = HTMLMediaElement.prototype.load;
  HTMLMediaElement.prototype.load = function() {
    originalLoad.call(this);
    this.pause();
    this.currentTime = 0;
  };

  // Watch for dynamically added media elements using MutationObserver
  const mediaObserver = new MutationObserver(mutations => {
    mutations.forEach(mutation => {
      mutation.addedNodes.forEach(node => {
        if (node.nodeType !== 1) return;

        if (node.nodeName === 'VIDEO' || node.nodeName === 'AUDIO') {
          freezeMedia(node);
        }

        if (node.querySelectorAll) {
          node.querySelectorAll('video, audio').forEach(freezeMedia);
        }
      });
    });
  });

  if (document.documentElement) {
    mediaObserver.observe(document.documentElement, {
      childList: true,
      subtree: true
    });
  } else {
    document.addEventListener('DOMContentLoaded', () => {
      mediaObserver.observe(document.documentElement, {
        childList: true,
        subtree: true
      });
    });
  }

  // Stub setInterval/setTimeout to prevent auto-advancing carousels
  const intervals = new Set();
  const timeouts = new Set();

  const originalSetInterval = window.setInterval;
  const originalSetTimeout = window.setTimeout;
  const originalClearInterval = window.clearInterval;
  const originalClearTimeout = window.clearTimeout;

  window.setInterval = function(...args) {
    const id = originalSetInterval.apply(this, args);
    intervals.add(id);
    return id;
  };

  window.setTimeout = function(...args) {
    const id = originalSetTimeout.apply(this, args);
    timeouts.add(id);
    return id;
  };

  window.clearInterval = function(id) {
    intervals.delete(id);
    return originalClearInterval(id);
  };

  window.clearTimeout = function(id) {
    timeouts.delete(id);
    return originalClearTimeout(id);
  };

  // Clear all intervals after page load and freeze all media elements
  window.addEventListener('load', () => {
    setTimeout(() => {
      intervals.forEach(id => originalClearInterval(id));
      intervals.clear();

      document.querySelectorAll('video, audio').forEach(freezeMedia);
    }, 100);
  });
})();
`

// FixIntersectionObserverScript makes all elements immediately visible.
const FixIntersectionObserverScript = `
(() => {
  window.IntersectionObserver = class IntersectionObserver {
    constructor(callback) {
      this.callback = callback;
    }

    observe(element) {
      setTimeout(() => {
        this.callback([{
          target: element,
          isIntersecting: true,
          intersectionRatio: 1.0,
          boundingClientRect: element.getBoundingClientRect(),
          intersectionRect: element.getBoundingClientRect(),
          rootBounds: null,
          time: Date.now()
        }], this);
      }, 0);
    }

    unobserve() {}
    disconnect() {}
    takeRecords() { return []; }
  };
})();
`

// DisableScrollScript disables scroll-related behaviors.
const DisableScrollScript = `
(() => {
  const noop = () => {};
  window.scrollTo = noop;
  window.scroll = noop;
  Element.prototype.scrollIntoView = function() {};
  Element.prototype.scrollTo = function() {};
  Element.prototype.scroll = function() {};
})();
`

// FreezeCarouselsScript stops and resets common carousel/slider libraries.
//
// NOTE: This is NOT a universal solution. It provides specific handling for
// popular slider libraries. Custom or less common slider implementations
// may not be affected by this script.
//
// Supported libraries:
//   - Swiper (swiper.js)
//   - Slick (slick.js, requires jQuery)
//   - Owl Carousel (owl.carousel.js, requires jQuery)
//   - Flickity (flickity.js)
//   - Bootstrap Carousel (bootstrap 5)
//
// For unsupported sliders, consider using --mask option to hide the element,
// or --inject-css to manually disable animations.
const FreezeCarouselsScript = `
(() => {
  // Freeze known slider library instances
  // All checks are defensive to avoid errors when libraries aren't present
  function freezeSliders() {
    // Swiper - check for .swiper property on elements
    document.querySelectorAll('.swiper-container, .swiper').forEach(el => {
      try {
        if (el.swiper && typeof el.swiper.slideTo === 'function') {
          if (el.swiper.autoplay && typeof el.swiper.autoplay.stop === 'function') {
            el.swiper.autoplay.stop();
          }
          el.swiper.slideTo(0, 0);
        }
      } catch (e) {}
    });

    // Slick - check for jQuery and slick plugin
    if (typeof jQuery !== 'undefined' && typeof jQuery.fn.slick === 'function') {
      try {
        jQuery('.slick-initialized').each(function() {
          jQuery(this).slick('slickPause');
          jQuery(this).slick('slickGoTo', 0, true);
        });
      } catch (e) {}
    }

    // Owl Carousel - check for jQuery and owlCarousel plugin
    if (typeof jQuery !== 'undefined' && typeof jQuery.fn.owlCarousel === 'function') {
      try {
        jQuery('.owl-carousel').trigger('stop.owl.autoplay');
        jQuery('.owl-carousel').trigger('to.owl.carousel', [0, 0]);
      } catch (e) {}
    }

    // Flickity - check for Flickity global and data method
    if (typeof Flickity !== 'undefined' && typeof Flickity.data === 'function') {
      document.querySelectorAll('.flickity-enabled').forEach(el => {
        try {
          const flkty = Flickity.data(el);
          if (flkty) {
            if (typeof flkty.pausePlayer === 'function') flkty.pausePlayer();
            if (typeof flkty.select === 'function') flkty.select(0, false, true);
          }
        } catch (e) {}
      });
    }

    // Bootstrap 5 Carousel - check for bootstrap global
    if (typeof bootstrap !== 'undefined' && bootstrap.Carousel && typeof bootstrap.Carousel.getInstance === 'function') {
      document.querySelectorAll('.carousel').forEach(el => {
        try {
          const carousel = bootstrap.Carousel.getInstance(el);
          if (carousel) {
            if (typeof carousel.pause === 'function') carousel.pause();
            if (typeof carousel.to === 'function') carousel.to(0);
          }
        } catch (e) {}
      });
    }

    // Generic: Reset transforms on common slider wrapper elements
    // This is safe even if the libraries aren't present
    ['.swiper-wrapper', '.slick-track', '.owl-stage', '.flickity-slider'].forEach(selector => {
      document.querySelectorAll(selector).forEach(el => {
        try {
          el.style.transform = 'translate3d(0, 0, 0)';
          el.style.transition = 'none';
        } catch (e) {}
      });
    });
  }

  // Run after page load (don't interfere with loading)
  window.addEventListener('load', () => {
    freezeSliders();
    setTimeout(freezeSliders, 100);
    setTimeout(freezeSliders, 300);
    setTimeout(freezeSliders, 500);
    setTimeout(freezeSliders, 1000);
  });

  // Expose for manual use
  window.__freezeSliders = freezeSliders;
})();
`

// DisableWebAnimationsScript disables the Web Animations API.
const DisableWebAnimationsScript = `
(() => {
  Element.prototype.animate = function() {
    return {
      cancel: () => {},
      finish: () => {},
      pause: () => {},
      play: () => {},
      reverse: () => {},
      playbackRate: 0,
      playState: 'finished',
      addEventListener: () => {},
      removeEventListener: () => {},
      dispatchEvent: () => false
    };
  };

  if (document.getAnimations) {
    document.getAnimations = () => [];
  }

  if (Element.prototype.getAnimations) {
    Element.prototype.getAnimations = () => [];
  }
})();
`

// GetAllDeterministicScripts combines all deterministic scripts.
// If mockTime is provided, the mock time script is included.
func GetAllDeterministicScripts(mockTime string) string {
	scripts := ""

	if mockTime != "" {
		scripts += GenerateMockTimeScript(mockTime) + "\n\n"
	}

	scripts += DisableAutoplayScript + "\n\n"
	scripts += FixIntersectionObserverScript + "\n\n"
	scripts += DisableScrollScript + "\n\n"
	scripts += DisableWebAnimationsScript + "\n\n"
	scripts += FreezeCarouselsScript

	return scripts
}
