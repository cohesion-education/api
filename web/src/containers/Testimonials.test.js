import { mount } from 'enzyme'
import { Provider } from 'react-redux'
import { Testimonials } from './'


jest.mock('../views/TestimonialList')

describe("<Testimonials /> Container ", () => {
    let wrapper
    let _store = {
        dispatch: jest.fn(),
        subscribe: jest.fn(),
        getState: jest.fn(() =>
          ({
              testimonials:[{
                text: 'Super Awesome!',
                avatar:'test-avatar.png',
                name:'Satisfied Parent'
              }]
          })
        )
    }

    beforeAll(() => wrapper = mount(
      <Provider store={_store}>
        <Testimonials />
      </Provider>
    ))

    afterEach(() => jest.resetAllMocks())

    it("renders TestimonialList with props", () => {
      expect(wrapper
        .find('TestimonialList')
        .props()
        .list
        .length
      ).toBe(1)
    })
})
