# http://guides.rubyonrails.org/v3.2.13/performance_testing.html

require 'test_helper'
require 'rails/performance_test_help'
require 'integration_helper'
require 'selenium-webdriver'
require 'headless'

class SearchTest < ActionDispatch::PerformanceTest
  self.profile_options = { :runs => 10,
                           :metrics => [:wall_time],
                           :output => 'tmp/performance',
                           :formats => [:flat] }

  setup do
    headless = Headless.new
    headless.start
    Capybara.current_driver = :selenium
    Capybara.current_session.driver.browser.manage.window.resize_to(1024, 768)
  end

  def test_search
    visit page_with_token('active')

    within('.navbar-fixed-top') do
      page.find_field('search').set('hash')
      page.find('.glyphicon-search').click
    end

    # We should now be in the search dialog. Expect at least one item in the result display.
    within '.modal-content' do
      assert_text 'All projects'
      assert_text 'Search'
      assert_selector('div', text: 'zzzzz-')
    end
  end
end
