#!/usr/bin/env ruby

require "net/http"
require "uri"
require "socket"
require "json"
require "openssl"

class String
  def is_json?
    begin
      !!JSON.parse(self)
    rescue
      false
    end ### begin
  end ### def
end ### class


def time_diff_milli(start, finish)
   (finish - start) * 1000.0
end

def url_test(uri)

  begin

    http    = Net::HTTP.new(uri.host, uri.port)
    http.open_timeout = 5
    http.read_timeout = 5
    if (uri.scheme == 'https') then
      http.use_ssl      = true
      http.verify_mode  = OpenSSL::SSL::VERIFY_NONE # read into this
    end ### if https
    request = Net::HTTP::Get.new(uri.request_uri)
    response = http.request(request)

    if response.body.is_json? then
      json = JSON.parse(response.body)
      return [response.code, response.body]
    else ### if json
      return ['not json', 'response is NOT json']
    end ### if json

  rescue Timeout::Error
    return ['timeout', 'timeout error']

  rescue Errno::ECONNREFUSED
    return ['connection refused', 'connection refused']

  rescue Errno::ECONNRESET
    return ['connected reset', 'Connection reset by peer']

  end ### begin

  return ['unknown', 'unknown problem']

end ### def


##
##
##
STATUS_OK                    = 0
STATUS_WARNING               = 1
STATUS_CRITICAL              = 2
STATUS_UNKNOWN               = 3

STATUS_OK_TXT                = "OK"
STATUS_WARNING_TXT           = "WARNING"
STATUS_CRITICAL_TXT          = "CRITICAL"
STATUS_UNKNOWN_TXT           = "WARNING"

azhealthcheck_status         = STATUS_OK
azhealthcheck_status_txt     = STATUS_OK_TXT
azhealthcheck_status_msg     = ""
azhealthcheck_status_longmsg = ""


azhealthcheck_url = 'http://0.0.0.0:3000/'

azhealthcheck_uri = URI.parse(azhealthcheck_url)

start_time = Time.now
code, body = url_test(azhealthcheck_uri)
stop_time  = Time.now
azhealthcheck_response_time = time_diff_milli(start_time, stop_time)

if ( code == "timeout" ) then
  azhealthcheck_status     = STATUS_CRITICAL
  azhealthcheck_status_txt = STATUS_CRITICAL_TXT
  azhealthcheck_status_msg = "Connection to AZ_Healthecheck TIMING OUT: #{azhealthcheck_url}, This could indicate that the AZ_Healthcheck service is not running."

elsif ( code == "not json" ) then
  azhealthcheck_status         = STATUS_CRITICAL
  azhealthcheck_status_txt     = STATUS_CRITICAL_TXT
  azhealthcheck_status_msg     = "AZ_Healthcheck NOT RETURNING JSON strings: #{azhealthcheck_url}, AZ_Healthcheck always responds with strings.. Something else may be responding on http:3000 or AZ_Healthcheck may be in an invalid state."
  azhealthcheck_status_longmsg = "#{azhealthcheck_status_longmsg}\\n\\n" + body.to_s.gsub(/\n/, "\\n")

elsif ( code.to_s == "connection refused") then
  azhealthcheck_status     = STATUS_CRITICAL
  azhealthcheck_status_txt = STATUS_CRITICAL_TXT
  azhealthcheck_status_msg = "AZ_Healthcheck CONNECTION REFUSED: #{azhealthcheck_url}, This could indicate a software firewall or SecurityGroup is in place blocking port 3000."

elsif ( code.to_s != '200' ) then
  azhealthcheck_status         = STATUS_WARNING
  azhealthcheck_status_txt     = STATUS_WARNING_TXT
  if ((JSON.parse(body))['statusText'] == 'unhealthy') then
    azhealthcheck_status_msg = azhealthcheck_status_msg + "UNHEALTHY :: "
  end
  azhealthcheck_status_msg     = azhealthcheck_status_msg + "returning HTTP " + code.to_s + "; See longmsg for the body output."
  azhealthcheck_status_longmsg = "#{azhealthcheck_status_longmsg}\\n\\n<h2>AZ_Healthcheck</h2>\\n\\n<pre>" + body.to_s.gsub(/\n/, "\\n") + "</pre>"
else
  azhealthcheck_status_msg     = "#{azhealthcheck_status_msg}; See longmsg for the body output."
  azhealthcheck_status_longmsg = "#{azhealthcheck_status_longmsg}\\n\\n<h2>AZ_Healthcheck</h2>\\n<pre>" + JSON.pretty_generate(JSON.parse(body.to_s)).gsub(/\n/, "\\n") + "</pre>" 
end



# Status (Nagios codes)
#   0 = OK
#   1 = WARNING
#   2 = CRITICAL
#   3 = UNKNOWN
# Item-name (underscore separated words)
# Performance-data;
#   varname=value;warn;crit;min;max|varname=value;warn;crit;min;max
# Check-output
puts azhealthcheck_status.to_s  + ' ' + 'az_healthcheck ' + 
     'azhealthcheck_response_time_milliseconds=' + azhealthcheck_response_time.to_s + ';400;600;1;800 ' + 
     azhealthcheck_status_txt  + ' - ' + 
     azhealthcheck_status_msg + ' \n\n\n' + 
     azhealthcheck_status_longmsg

