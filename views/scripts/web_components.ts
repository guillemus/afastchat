import { rocket } from '/public/datastar-rocket-v1.0.4.js'

rocket('sensitive-text', {
    mode: 'light',
    setup: ({ $$ }) => {
        $$('revealed', false)
    },
})
