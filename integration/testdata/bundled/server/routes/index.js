import { eventHandler } from "h3"
import "leftpad"

export default eventHandler((event) => {
  return "hello world";
});